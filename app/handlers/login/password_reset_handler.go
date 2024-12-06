package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/cptutil"
	"github.com/askasoft/pango-xdemo/app/utils/smtputil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/vad"
	"github.com/askasoft/pango/xin"
)

type PwdRstToken struct {
	Email     string `json:"email"`
	Timestamp int64  `json:"timestamp"`
}

func (pts *PwdRstToken) String() string {
	bs, err := json.Marshal(pts)
	if err != nil {
		panic(err)
	}
	return str.UnsafeString(bs)
}

type PwdRstArg struct {
	Newpwd string `form:"newpwd" validate:"required,printascii"`
	Conpwd string `form:"conpwd" validate:"required,eqfield=Newpwd"`
}

func PasswordResetIndex(c *xin.Context) {
	h := handlers.H(c)

	c.HTML(http.StatusOK, "login/pwdrst_mail", h)
}

func PasswordResetSend(c *xin.Context) {
	email := str.Strip(c.PostForm("email"))

	if email == "" || !vad.IsEmail(email) {
		c.AddError(&vadutil.ParamError{Param: "email", Message: tbs.GetText(c.Locale, "pwdrst.error.email")})
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	user, err := tenant.FindUser(c, email)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	if user == nil {
		// show success message even user is not found
		c.JSON(http.StatusOK, xin.H{
			"success": tbs.Format(c.Locale, "pwdrst.success.sent", email),
		})
		return
	}

	token := &PwdRstToken{Email: email, Timestamp: time.Now().UnixMilli()}
	tkenc := cptutil.MustEncrypt(app.Secret(), token.String())
	rsurl := fmt.Sprintf("%s://%s%s/login/pwdrst/reset/%s", str.If(c.IsSecure(), "https", "http"), c.RequestHostname(), app.Base, tkenc)

	tkexp := num.Itoa(int(app.INI.GetDuration("login", "passwordResetTokenExpires", time.Minute*10).Minutes()))

	sr := strings.NewReplacer(
		"{{SITE_NAME}}", tbs.GetText(c.Locale, "title"),
		"{{USER_NAME}}", user.(*models.User).Name,
		"{{USER_EMAIL}}", user.(*models.User).Email,
		"{{REQUEST_DATE}}", app.FormatTime(time.Now()),
		"{{RESET_URL}}", rsurl,
		"{{EXPIRES}}", tkexp,
	)
	subject := sr.Replace(tbs.GetText(c.Locale, "pwdrst.email.send.subject"))

	sr = strings.NewReplacer(
		"{{SITE_NAME}}", html.EscapeString(tbs.GetText(c.Locale, "title")),
		"{{USER_NAME}}", html.EscapeString(user.(*models.User).Name),
		"{{USER_EMAIL}}", html.EscapeString(user.(*models.User).Email),
		"{{REQUEST_DATE}}", html.EscapeString(app.FormatTime(time.Now())),
		"{{RESET_URL}}", rsurl,
		"{{EXPIRES}}", tkexp,
	)
	message := sr.Replace(tbs.GetText(c.Locale, "pwdrst.email.send.message"))

	if err := smtputil.SendHTMLMail(email, subject, message); err != nil {
		c.Logger.Error(err)
		c.AddError(errors.New(tbs.GetText(c.Locale, "pwdrst.error.sendmail")))
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "pwdrst.success.sent", email),
	})
}

func passwordResetToken(c *xin.Context) *PwdRstToken {
	tkenc := c.Param("token")
	tkstr, err := cptutil.Decrypt(app.Secret(), tkenc)
	if tkenc == "" || err != nil {
		c.AddError(errors.New(tbs.GetText(c.Locale, "pwdrst.error.invalid")))
		return nil
	}

	token := &PwdRstToken{}
	if err := json.Unmarshal(str.UnsafeBytes(tkstr), token); err != nil {
		c.AddError(errors.New(tbs.GetText(c.Locale, "pwdrst.error.invalid")))
		return nil
	}

	tkexp := app.INI.GetDuration("login", "passwordResetTokenExpires", time.Minute*10)
	tktm := time.UnixMilli(token.Timestamp)
	if time.Since(tktm) > tkexp {
		c.AddError(errors.New(tbs.GetText(c.Locale, "pwdrst.error.expired")))
		return nil
	}

	return token
}

func PasswordResetConfirm(c *xin.Context) {
	token := passwordResetToken(c)

	h := handlers.H(c)
	if token != nil {
		h["Message"] = tbs.Format(c.Locale, "pwdrst.confirm", token.Email)
	}

	c.HTML(http.StatusOK, "login/pwdrst_exec", h)
}

func PasswordResetExecute(c *xin.Context) {
	token := passwordResetToken(c)
	if token == nil {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	user, err := tenant.FindUser(c, token.Email)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}
	if user == nil {
		c.AddError(errors.New(tbs.GetText(c.Locale, "pwdrst.error.invalid")))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	pra := &PwdRstArg{}
	if err := c.Bind(pra); err != nil {
		vadutil.AddBindErrors(c, err, "pwdrst.")
	}

	if pra.Newpwd != "" {
		if vs := tt.ValidatePassword(c.Locale, pra.Newpwd); len(vs) > 0 {
			for _, v := range vs {
				c.AddError(&vadutil.ParamError{
					Param:   "newpwd",
					Message: v,
				})
			}
		}
	}

	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	mu := user.(*models.User)
	mu.SetPassword(pra.Newpwd)

	sqb := app.SDB.Builder()
	sqb.Update(tt.TableUsers())
	sqb.Setc("password", mu.Password)
	sqb.Where("id = ?", mu.ID)
	sql, args := sqb.Build()

	if _, err := app.SDB.Exec(sql, args...); err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	sr := strings.NewReplacer(
		"{{SITE_NAME}}", tbs.GetText(c.Locale, "title"),
		"{{USER_NAME}}", user.(*models.User).Name,
		"{{USER_EMAIL}}", user.(*models.User).Email,
		"{{RESET_DATE}}", app.FormatTime(time.Now()),
	)
	subject := sr.Replace(tbs.GetText(c.Locale, "pwdrst.email.reset.subject"))

	sr = strings.NewReplacer(
		"{{SITE_NAME}}", html.EscapeString(tbs.GetText(c.Locale, "title")),
		"{{USER_NAME}}", html.EscapeString(user.(*models.User).Name),
		"{{USER_EMAIL}}", html.EscapeString(user.(*models.User).Email),
		"{{RESET_DATE}}", html.EscapeString(app.FormatTime(time.Now())),
	)
	message := sr.Replace(tbs.GetText(c.Locale, "pwdrst.email.reset.message"))

	if err := smtputil.SendHTMLMail(token.Email, subject, message); err != nil {
		c.Logger.Error(err)
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "pwdrst.success.reset", token.Email),
	})
}
