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
	uq, err := userListArgs(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	c.SetAttachmentHeader("users.csv")

	_, _ = c.Writer.WriteString(string(iox.BOM))

	cw := csv.NewWriter(c.Writer)
	cw.UseCRLF = true

	tt := tenant.FromCtx(c)

	sqb := app.SDB.Builder()
	sqb.Select()
	sqb.From(tt.TableUsers())
	uq.AddWhere(c, sqb)
	sqb.Order("id")
	sql, args := sqb.Build()

	rows, err := app.SDB.Queryx(sql, args...)
	if err != nil {
		c.Logger.Error(err)
		_ = cw.Write([]string{err.Error()})
		cw.Flush()
		return
	}
	defer rows.Close()

	err = cw.Write([]string{
		tbs.GetText(c.Locale, "user.id"),
		tbs.GetText(c.Locale, "user.name"),
		tbs.GetText(c.Locale, "user.email"),
		tbs.GetText(c.Locale, "user.role"),
		tbs.GetText(c.Locale, "user.status"),
		tbs.GetText(c.Locale, "user.password"),
		tbs.GetText(c.Locale, "user.cidr"),
		tbs.GetText(c.Locale, "user.created_at"),
		tbs.GetText(c.Locale, "user.updated_at"),
	})
	if err != nil {
		c.Logger.Error(err)
		return
	}

	au := tenant.AuthUser(c)
	sm := tbsutil.GetUserStatusMap(c.Locale)
	rm := tbsutil.GetUserRoleMap(c.Locale, au.Role)
	for rows.Next() {
		var user models.User
		err = rows.StructScan(&user)
		if err != nil {
			c.Logger.Error(err)
			_ = cw.Write([]string{err.Error()})
			cw.Flush()
			return
		}

		err = cw.Write([]string{
			num.Ltoa(user.ID),
			user.Name,
			user.Email,
			rm.MustGet(user.Role),
			sm.MustGet(user.Status, user.Status),
			"",
			user.CIDR,
			models.FormatTime(user.CreatedAt),
			models.FormatTime(user.UpdatedAt),
		})
		if err != nil {
			c.Logger.Error(err)
			return
		}
	}

	cw.Flush()
}
