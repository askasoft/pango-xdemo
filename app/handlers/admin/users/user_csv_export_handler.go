package users

import (
	"encoding/csv"
	"net/http"

	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/args"
	"github.com/askasoft/pangox-xdemo/app/middles"
	"github.com/askasoft/pangox-xdemo/app/models"
	"github.com/askasoft/pangox-xdemo/app/tenant"
	"github.com/askasoft/pangox-xdemo/app/utils/tbsutil"
)

func UserCsvExport(c *xin.Context) {
	uqa, err := bindUserQueryArg(c)
	if err != nil {
		args.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, middles.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	rm := tbsutil.GetUserRoleMap(c.Locale, au.Role)
	sm := tbsutil.GetUserStatusMap(c.Locale)
	mm := tbsutil.GetUserLoginMFAMap(c.Locale)

	cw := csv.NewWriter(c.Writer)
	cw.UseCRLF = true
	defer cw.Flush()

	var cols []string
	err = tt.IterUsers(app.SDB, au.Role, uqa, func(user *models.User) error {
		if len(cols) == 0 {
			c.SetAttachmentHeader("users.csv")
			_, _ = c.Writer.WriteString(string(iox.BOM))

			cols = append(cols,
				tbs.GetText(c.Locale, "user.id"),
				tbs.GetText(c.Locale, "user.name"),
				tbs.GetText(c.Locale, "user.email"),
				tbs.GetText(c.Locale, "user.password"),
				tbs.GetText(c.Locale, "user.role"),
				tbs.GetText(c.Locale, "user.status"),
				tbs.GetText(c.Locale, "user.login_mfa"),
				tbs.GetText(c.Locale, "user.cidr"),
				tbs.GetText(c.Locale, "user.created_at"),
				tbs.GetText(c.Locale, "user.updated_at"),
			)
			if err := cw.Write(cols); err != nil {
				return err
			}
		}

		cols = cols[:0]
		cols = append(cols,
			num.Ltoa(user.ID),
			user.Name,
			user.Email,
			"",
			rm.SafeGet(user.Role, user.Role),
			sm.SafeGet(user.Status, user.Status),
			mm.SafeGet(user.LoginMFA, user.LoginMFA),
			user.CIDR,
			app.FormatTime(user.CreatedAt),
			app.FormatTime(user.UpdatedAt),
		)
		return cw.Write(cols)
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, middles.E(c))
		return
	}
}
