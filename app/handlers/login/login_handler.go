package login

import (
	"encoding/base64"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/cptutil"
	"github.com/askasoft/pango-xdemo/app/utils/otputil"
	"github.com/askasoft/pango-xdemo/app/utils/smtputil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/ini"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/ran"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xmw"
	"github.com/skip2/go-qrcode"
	"github.com/xlzd/gotp"
)

func Index(c *xin.Context) {
	h := handlers.H(c)

	h["origin"] = c.Query(xmw.AuthRedirectOriginURLQuery)

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

func loginFindUser(c *xin.Context) (up *UserPass, au *models.User, ok bool) {
	up = &UserPass{}
	err := c.Bind(up)
	if err != nil {
		vadutil.AddBindErrors(c, err, "login.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if tenant.IsClientBlocked(c) {
		c.AddError(tbs.Error(c.Locale, "login.failed.blocked"))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	au, err = tt.FindAuthUser(up.Username)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	if au == nil || au.GetPassword() != up.Password {
		loginFailed(c, "login.failed.userpass")
		return
	}

	if !au.HasRole(models.RoleViewer) {
		loginFailed(c, "login.failed.notallowed")
		return
	}

	if !tenant.CheckClientIP(c, au) {
		loginFailed(c, "login.failed.restricted")
		return
	}

	ok = true
	return
}

func loginFailed(c *xin.Context, reason string) {
	tenant.AuthFailed(c)
	c.AddError(tbs.Error(c.Locale, reason))
	c.JSON(http.StatusBadRequest, handlers.E(c))
}

func loginPassed(c *xin.Context, au *models.User) {
	tt := tenant.FromCtx(c)
	if err := tt.Schema.AddAuditLog(app.SDB, au.ID, c.ClientIP(), models.AL_LOGIN_LOGIN, au.Email); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
	}

	if err := app.XCA.SaveUserPassToCookie(c, au.Email, au.GetPassword()); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	tenant.AuthPassed(c)

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.GetText(c.Locale, "login.success.loggedin"),
	})
}

func loginMFASecret(c *xin.Context, au *models.User) string {
	return cptutil.MustEncrypt(app.Secret(), c.RequestHostname()+"/"+au.Email+"/"+num.Ltoa(au.Secret))
}

func loginMFACheck(c *xin.Context, au *models.User, up *UserPass) bool {
	tt := tenant.FromCtx(c)
	mfa := tt.ConfigValue("secure_login_mfa")

	switch mfa {
	case "E":
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
	case "M":
		if up.Passcode == nil {
			c.JSON(http.StatusOK, xin.H{
				"message": tbs.GetText(c.Locale, "login.mfa.mobile.notice"),
				"mfa":     "M",
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
	sr := strings.NewReplacer(
		"{{SITE_NAME}}", tbs.GetText(c.Locale, "title"),
	)
	subject := sr.Replace(tbs.GetText(c.Locale, "login.mfa.email.subject"))

	sr = strings.NewReplacer(
		"{{SITE_NAME}}", html.EscapeString(tbs.GetText(c.Locale, "title")),
		"{{USER_EMAIL}}", html.EscapeString(email),
		"{{REQUEST_DATE}}", html.EscapeString(app.FormatTime(time.Now())),
		"{{PASSCODE}}", html.EscapeString(passcode),
		"{{EXPIRES}}", num.Itoa(int(expire.Minutes())),
	)
	message := sr.Replace(tbs.GetText(c.Locale, "login.mfa.email.message"))

	if err := smtputil.SendHTMLMail(email, subject, message); err != nil {
		c.Logger.Error(err)
		c.AddError(tbs.Error(c.Locale, "login.error.sendmail"))
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"message": tbs.Format(c.Locale, "login.mfa.email.notice", email),
		"mfa":     "E",
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
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	secret := loginMFASecret(c, au)
	expire := ini.GetDuration("login", "mobilePasscodeExpires", 30*time.Second)
	totp := otputil.NewTOTP(secret, expire)
	loginSendEmailQrcode(c, au.Email, totp)
}

func loginSendEmailQrcode(c *xin.Context, email string, totp *gotp.TOTP) {
	sr := strings.NewReplacer(
		"{{SITE_NAME}}", tbs.GetText(c.Locale, "title"),
	)
	subject := sr.Replace(tbs.GetText(c.Locale, "login.mfa.mobile.subject"))

	purl := totp.ProvisioningUri(email, c.Request.Host)
	png, err := qrcode.Encode(purl, qrcode.Medium, 256)
	if err != nil {
		c.Logger.Error(err)
		c.AddError(tbs.Error(c.Locale, "login.error.qrcode"))
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	sr = strings.NewReplacer(
		"{{SITE_NAME}}", html.EscapeString(tbs.GetText(c.Locale, "title")),
		"{{USER_EMAIL}}", html.EscapeString(email),
		"{{QRCODE}}", base64.StdEncoding.EncodeToString(png),
	)
	message := sr.Replace(tbs.GetText(c.Locale, "login.mfa.mobile.message"))

	if err := smtputil.SendHTMLMail(email, subject, message); err != nil {
		c.Logger.Error(err)
		c.AddError(tbs.Error(c.Locale, "login.error.sendmail"))
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "login.mfa.mobile.qrcsent", email),
	})
}

func Logout(c *xin.Context) {
	tenant.DeleteAuthUser(c)

	app.XCA.DeleteCookie(c)

	h := handlers.H(c)
	h["Message"] = tbs.GetText(c.Locale, "login.success.loggedout")

	c.HTML(http.StatusOK, "login/login", h)
}
