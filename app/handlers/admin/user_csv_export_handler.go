package admin

import (
	"encoding/csv"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func UserCsvExport(c *xin.Context) {
	q, err := userListArgs(c)
	if err != nil {
		utils.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	c.SetAttachmentHeader("users.csv")

	_, _ = c.Writer.WriteString(string(iox.BOM))

	cw := csv.NewWriter(c.Writer)
	cw.UseCRLF = true

	tt := tenant.FromCtx(c)

	tx := app.GDB.Table(tt.TableUsers())
	tx = filterUsers(c)(tx, q).Order("id ASC")

	rows, err := tx.Rows()
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
		tbs.GetText(c.Locale, "user.cidr"),
		tbs.GetText(c.Locale, "user.created_at"),
		tbs.GetText(c.Locale, "user.updated_at"),
	})
	if err != nil {
		c.Logger.Error(err)
		return
	}

	sm := utils.GetUserStatusMap(c.Locale)
	rm := utils.GetUserRoleMap(c.Locale)
	for rows.Next() {
		var usr models.User
		err = tx.ScanRows(rows, &usr)
		if err != nil {
			_ = cw.Write([]string{err.Error()})
			cw.Flush()
			return
		}

		err = cw.Write([]string{
			num.Ltoa(usr.ID),
			usr.Name,
			usr.Email,
			rm.MustGet(usr.Role, usr.Role),
			sm.MustGet(usr.Status, usr.Status),
			usr.CIDR,
			models.FormatTime(usr.CreatedAt),
			models.FormatTime(usr.UpdatedAt),
		})
		if err != nil {
			c.Logger.Error(err)
			return
		}
	}

	cw.Flush()
}
