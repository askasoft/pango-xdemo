package login

import (
	"errors"
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
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pango/xmw"
)

func Index(c *xin.Context) {
	h := handlers.H(c)

	h["origin"] = c.Query(xmw.AuthRedirectOriginURLQuery)

	c.HTML(http.StatusOK, "login/login", h)
}

type UserPass struct {
	Username string `form:"username" validate:"required"`
	Password string `form:"password" validate:"required"`
	Passcode string `form:"passcode"`
}

func Login(c *xin.Context) {
	userpass := &UserPass{}
	if err := c.Bind(userpass); err != nil {
		vadutil.AddBindErrors(c, err, "login.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if tenant.IsClientBlocked(c) {
		c.AddError(errors.New(tbs.GetText(c.Locale, "login.failed.blocked")))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	au, err := tenant.FindUser(c, userpass.Username)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	reason := "login.failed.userpass"

	if au != nil && userpass.Password == au.GetPassword() {
		user := au.(*models.User)
		if user.HasRole(models.RoleViewer) {
			if tenant.CheckClientIP(c, user) {
				if loginMFACheck(c, userpass) {
					err := app.XCA.SaveUserPassToCookie(c, userpass.Username, userpass.Password)
					if err != nil {
						c.AddError(err)
						c.JSON(http.StatusInternalServerError, handlers.E(c))
						return
					}

					tenant.AuthPassed(c)

					c.JSON(http.StatusOK, xin.H{
						"success": tbs.GetText(c.Locale, "login.success.loggedin"),
					})
				}
				return
			}
			reason = "login.failed.restricted"
		} else {
			reason = "login.failed.notallowed"
		}
	}

	tenant.AuthFailed(c)
	c.AddError(errors.New(tbs.GetText(c.Locale, reason)))
	c.JSON(http.StatusBadRequest, handlers.E(c))
}

func loginMFACheck(c *xin.Context, userpass *UserPass) bool {
	tt := tenant.FromCtx(c)
	mfa := tt.ConfigValue("secure_login_mfa")
	switch mfa {
	case "E":
		secret := cptutil.Hash(userpass.Username)
		expire := app.INI.GetDuration("login", "emailPasscodeExpires", 10*time.Minute)
		totp := otputil.NewTOTP(secret, expire)
		if userpass.Passcode == "" {
			loginSendEmailPasscode(c, userpass.Username, totp.Now(), expire)
			return false
		}

		if otputil.TOTPVerify(totp, expire, userpass.Passcode) {
			return true
		}

		c.AddError(errors.New(tbs.GetText(c.Locale, "login.failed.passcode")))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return false
	case "A":
		return true
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
		"{{REQUEST_DATE}}", html.EscapeString(models.FormatTime(time.Now())),
		"{{PASSCODE}}", html.EscapeString(passcode),
		"{{EXPIRES}}", num.Itoa(int(expire.Minutes())),
	)
	message := sr.Replace(tbs.GetText(c.Locale, "login.mfa.email.message"))

	if err := smtputil.SendHTMLMail(email, subject, message); err != nil {
		c.Logger.Error(err)
		c.AddError(errors.New(tbs.GetText(c.Locale, "login.error.sendmail")))
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"mfa":     true,
		"message": tbs.Format(c.Locale, "login.mfa.email.sent", email),
	})
}

func Logout(c *xin.Context) {
	tenant.DeleteAuthUser(c)

	app.XCA.DeleteCookie(c)

	h := handlers.H(c)
	h["Message"] = tbs.GetText(c.Locale, "login.success.loggedout")

	c.HTML(http.StatusOK, "login/login", h)
}
