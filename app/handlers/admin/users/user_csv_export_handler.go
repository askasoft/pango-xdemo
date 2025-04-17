package users

import (
	"encoding/csv"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func UserCsvExport(c *xin.Context) {
	uqa, err := bindUserQueryArg(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	sm := tbsutil.GetUserStatusMap(c.Locale)
	rm := tbsutil.GetUserRoleMap(c.Locale, au.Role)

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
				tbs.GetText(c.Locale, "user.role"),
				tbs.GetText(c.Locale, "user.status"),
				tbs.GetText(c.Locale, "user.password"),
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
			rm.SafeGet(user.Role, user.Role),
			sm.SafeGet(user.Status, user.Status),
			"",
			user.CIDR,
			app.FormatTime(user.CreatedAt),
			app.FormatTime(user.UpdatedAt),
		)
		return cw.Write(cols)
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}
}
