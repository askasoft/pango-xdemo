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

func PasswordResetIndex(c *xin.Context) {
	h := handlers.H(c)

	c.HTML(http.StatusOK, "login/pwdrst", h)
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
		c.AddError(&vadutil.ParamError{Param: "email", Message: tbs.GetText(c.Locale, "pwdrst.error.email")})
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	token := &PwdRstToken{Email: email, Timestamp: time.Now().UnixMilli()}
	tkenc := cptutil.MustEncrypt(app.Secret(), token.String())
	rsurl := fmt.Sprintf("%s://%s%s/login/pwdrst/reset/%s", str.If(c.IsSecure(), "https", "http"), c.Request.Host, app.Base, tkenc)

	sr := strings.NewReplacer(
		"{{SITE_NAME}}", tbs.GetText(c.Locale, "title"),
		"{{USER_NAME}}", user.(*models.User).Name,
		"{{REQUEST_DATE}}", models.FormatTime(time.Now()),
		"{{RESET_URL}}", rsurl,
	)
	subject := sr.Replace(tbs.GetText(c.Locale, "pwdrst.email.send.subject"))

	sr = strings.NewReplacer(
		"{{SITE_NAME}}", html.EscapeString(tbs.GetText(c.Locale, "title")),
		"{{USER_NAME}}", html.EscapeString(user.(*models.User).Name),
		"{{REQUEST_DATE}}", html.EscapeString(models.FormatTime(time.Now())),
		"{{RESET_URL}}", rsurl,
	)
	message := sr.Replace(tbs.GetText(c.Locale, "pwdrst.email.send.message"))

	if err := smtputil.SendHTMLMail(email, subject, message); err != nil {
		c.Logger.Error(err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{"success": tbs.Format(c.Locale, "pwdrst.success.sent", email)})
}

func PasswordResetReset(c *xin.Context) {
	tkenc := c.Param("token")
	tkstr, err := cptutil.Decrypt(app.Secret(), tkenc)
	if tkenc == "" || err != nil {
		c.AddError(errors.New(tbs.GetText(c.Locale, "pwdrst.error.link")))
		c.HTML(http.StatusOK, "login/pwdrst_reset", handlers.H(c))
		return
	}

	token := &PwdRstToken{}
	if err := json.Unmarshal(str.UnsafeBytes(tkstr), token); err != nil {
		c.AddError(errors.New(tbs.GetText(c.Locale, "pwdrst.error.link")))
		c.HTML(http.StatusOK, "login/pwdrst_reset", handlers.H(c))
		return
	}

	tkexp := app.INI.GetDuration("login", "passwordResetTokenExpires", time.Hour*2)
	tktm := time.UnixMicro(token.Timestamp)
	if tktm.Add(tkexp).After(time.Now()) {
		c.AddError(errors.New(tbs.GetText(c.Locale, "pwdrst.error.timeout")))
		c.HTML(http.StatusOK, "login/pwdrst_reset", handlers.H(c))
		return
	}

	user, err := tenant.FindUser(c, token.Email)
	if err != nil {
		c.Logger.Error(err)
		c.AddError(err)
		c.HTML(http.StatusOK, "login/pwdrst_reset", handlers.H(c))
		return
	}
	if user == nil {
		c.AddError(errors.New(tbs.GetText(c.Locale, "pwdrst.error.link")))
		c.HTML(http.StatusOK, "login/pwdrst_reset", handlers.H(c))
		return
	}

	tt := tenant.FromCtx(c)

	password := str.RandLetterNumbers(16)
	mu := user.(*models.User)
	mu.SetPassword(password)

	if err := app.GDB.Table(tt.TableUsers()).Where("id = ?", mu.ID).Update("password", mu.Password).Error; err != nil {
		c.Logger.Error(err)
		c.AddError(err)
		c.HTML(http.StatusOK, "login/pwdrst_reset", handlers.H(c))
		return
	}

	sr := strings.NewReplacer(
		"{{SITE_NAME}}", tbs.GetText(c.Locale, "title"),
		"{{USER_NAME}}", user.(*models.User).Name,
		"{{PASSWORD}}", password,
	)
	subject := sr.Replace(tbs.GetText(c.Locale, "pwdrst.email.reset.subject"))

	sr = strings.NewReplacer(
		"{{SITE_NAME}}", html.EscapeString(tbs.GetText(c.Locale, "title")),
		"{{USER_NAME}}", html.EscapeString(user.(*models.User).Name),
		"{{PASSWORD}}", html.EscapeString(password),
	)
	message := sr.Replace(tbs.GetText(c.Locale, "pwdrst.email.reset.message"))

	if err := smtputil.SendHTMLMail(token.Email, subject, message); err != nil {
		c.Logger.Error(err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)
	h["Success"] = tbs.Format(c.Locale, "pwdrst.success.reset", token.Email)
	c.HTML(http.StatusOK, "login/pwdrst_reset", h)
}
