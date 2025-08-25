package login

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/ran"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xin/middleware"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/middles"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox-xdemo/app/utils/otputil"
	"github.com/askasoft/pangox-xdemo/app/utils/smtputil"
	"github.com/askasoft/pangox/xwa/xcpts"
	"github.com/skip2/go-qrcode"
	"github.com/xlzd/gotp"
)

func Index(c *xin.Context) {
	h := middles.H(c)

	h["origin"] = c.Query(middleware.AuthRedirectOriginURLQuery)

	c.HTML(http.StatusOK, "login/login", h)
}

type UserPass struct {
	Username string  `form:"username" validate:"required"`
	Password string  `form:"password" validate:"required"`
	Passcode *string `form:"passcode"`
}

func Login(c *xin.Context) {
	up, au, ok := loginFindUser(c)
	if !ok {
		return
	}

	if loginMFACheck(c, au, up) {
		loginPassed(c, au)
	}
}

func loginFindUser(c *xin.Context) (up *UserPass, mu *models.User, ok bool) {
	up = &UserPass{}
	err := c.Bind(up)
	if err != nil {
		args.AddBindErrors(c, err, "login.")
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	if tenant.IsClientBlocked(c) {
		c.AddError(tbs.Error(c.Locale, "login.failed.blocked"))
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	au, err := tenant.Authenticate(c, up.Username, up.Password)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	if au == nil {
		loginFailed(c, "login.failed.userpass")
		return
	}

	mu = au.(*models.User)
	if !mu.HasRole(models.RoleViewer) {
		loginFailed(c, "login.failed.notallowed")
		return
	}

	if !tenant.CheckUserClientIP(c, mu) {
		loginFailed(c, "login.failed.restricted")
		return
	}

	ok = true
	return
}

func loginFailed(c *xin.Context, reason string) {
	tenant.AuthFailed(c)
	c.AddError(tbs.Error(c.Locale, reason))
	c.JSON(http.StatusBadRequest, middles.E(c))
}

func loginPassed(c *xin.Context, au *models.User) {
	tt := tenant.FromCtx(c)
	if err := tt.Schema.AddAuditLog(app.SDB, au.ID, c.ClientIP(), au.Role, models.AL_LOGIN_LOGIN, au.Email); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
	}

	if err := app.XCA.SaveUserPassToCookie(c, au); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	tenant.AuthPassed(c)

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.GetText(c.Locale, "login.success.loggedin"),
	})
}

func loginMFASecret(c *xin.Context, au *models.User) string {
	return xcpts.MustEncrypt(app.Secret(), c.RequestHostname()+"/"+au.Email+"/"+num.Ltoa(au.Secret))
}

func loginMFACheck(c *xin.Context, au *models.User, up *UserPass) bool {
	mfa := au.LoginMFA
	if mfa == app.LOGIN_MFA_UNSET {
		tt := tenant.FromCtx(c)
		mfa = tt.ConfigValue("secure_login_mfa")
	}

	switch mfa {
	case app.LOGIN_MFA_EMAIL:
		secret := loginMFASecret(c, au)
		expire := ini.GetDuration("login", "emailPasscodeExpires", 10*time.Minute)
		totp := otputil.NewTOTP(secret, expire)
		if up.Passcode == nil {
			loginSendEmailPasscode(c, au.Email, totp.Now(), expire)
			return false
		}

		if otputil.TOTPVerify(totp, expire, *up.Passcode) {
			return true
		}

		loginFailed(c, "login.failed.passcode")
		return false
	case app.LOGIN_MFA_MOBILE:
		if up.Passcode == nil {
			c.JSON(http.StatusOK, xin.H{
				"message": tbs.GetText(c.Locale, "login.mfa.mobile.notice"),
				"mfa":     app.LOGIN_MFA_MOBILE,
			})
			return false
		}

		secret := loginMFASecret(c, au)
		expire := ini.GetDuration("login", "mobilePasscodeExpires", 30*time.Second)
		totp := otputil.NewTOTP(secret, expire)

		if otputil.TOTPVerify(totp, expire, *up.Passcode) {
			return true
		}

		loginFailed(c, "login.failed.passcode")
		return false
	default:
		return true
	}
}

func loginSendEmailPasscode(c *xin.Context, email, passcode string, expire time.Duration) {
	h := middles.H(c)
	h["Email"] = email
	h["Passcode"] = passcode
	h["Expires"] = int(expire.Minutes())

	if err := smtputil.SendTemplateEmail(c.Locale, "email/login/passcode_send", email, h); err != nil {
		c.Logger.Error(err)
		c.AddError(tbs.Error(c.Locale, "login.error.sendmail"))
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"message": tbs.Format(c.Locale, "login.mfa.email.notice", email),
		"mfa":     app.LOGIN_MFA_EMAIL,
	})
}

func LoginMFAEnroll(c *xin.Context) {
	_, au, ok := loginFindUser(c)
	if !ok {
		return
	}

	au.Secret = ran.RandInt63()

	tt := tenant.FromCtx(c)

	_, err := tt.UpdateUserSecret(app.SDB, au.ID, au.Secret)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	secret := loginMFASecret(c, au)
	expire := ini.GetDuration("login", "mobilePasscodeExpires", 30*time.Second)
	totp := otputil.NewTOTP(secret, expire)
	loginSendEmailQrcode(c, au.Email, totp)
}

func loginSendEmailQrcode(c *xin.Context, email string, totp *gotp.TOTP) {
	purl := totp.ProvisioningUri(email, c.Request.Host)
	png, err := qrcode.Encode(purl, qrcode.Medium, 256)
	if err != nil {
		c.Logger.Error(err)
		c.AddError(tbs.Error(c.Locale, "login.error.qrcode"))
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	h := middles.H(c)
	h["Email"] = email
	h["QRCode"] = base64.StdEncoding.EncodeToString(png)

	if err := smtputil.SendTemplateEmail(c.Locale, "email/login/mbenroll_send", email, h); err != nil {
		c.Logger.Error(err)
		c.AddError(tbs.Error(c.Locale, "login.error.sendmail"))
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "login.mfa.mobile.qrcsent", email),
	})
}

func Logout(c *xin.Context) {
	tenant.DeleteAuthUser(c)

	app.XCA.DeleteCookie(c)

	h := middles.H(c)
	h["Message"] = tbs.GetText(c.Locale, "login.success.loggedout")

	c.HTML(http.StatusOK, "login/login", h)
}
