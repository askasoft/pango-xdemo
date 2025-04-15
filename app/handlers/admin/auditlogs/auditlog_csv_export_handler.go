package auditlogs

import (
	"encoding/csv"
	"net/http"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/asg"
	"github.com/askasoft/pango/iox"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

func AuditLogCsvExport(c *xin.Context) {
	uqa, err := bindAuditLogQueryArg(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	db := app.SDB

	sqb := db.Builder()
	sqb.Select("audit_logs.*", "COALESCE(users.email, '') AS user")
	sqb.From(tt.TableAuditLogs())
	sqb.Join("LEFT JOIN " + tt.TableUsers() + " ON users.id = audit_logs.uid")
	uqa.AddFilters(c, sqb)
	sqb.Order("id")
	sql, args := sqb.Build()

	rows, err := db.Queryx(sql, args...)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}
	defer rows.Close()

	c.SetAttachmentHeader("auditlogs.csv")
	_, _ = c.Writer.WriteString(string(iox.BOM))

	cw := csv.NewWriter(c.Writer)
	cw.UseCRLF = true
	defer cw.Flush()

	cols := []string{
		tbs.GetText(c.Locale, "auditlog.id"),
		tbs.GetText(c.Locale, "auditlog.date"),
		tbs.GetText(c.Locale, "auditlog.user"),
		tbs.GetText(c.Locale, "auditlog.func"),
		tbs.GetText(c.Locale, "auditlog.action"),
		tbs.GetText(c.Locale, "auditlog.message"),
	}
	if err = cw.Write(cols); err != nil {
		c.Logger.Error(err)
		return
	}

	fm := tbsutil.GetAudioLogFuncMap(c.Locale)
	for rows.Next() {
		var al models.AuditLogEx
		if err = rows.StructScan(&al); err != nil {
			c.Logger.Error(err)
			_ = cw.Write([]string{err.Error()})
			return
		}

		if len(al.Params) > 0 {
			al.Detail = tbs.Format(c.Locale, "auditlog.detail."+al.Func+"."+al.Action, asg.Anys(al.Params)...)
		}
		cols = []string{
			num.Ltoa(al.ID),
			app.FormatTime(al.Date),
			al.User,
			fm.SafeGet(al.Func, al.Func),
			tbs.Format(c.Locale, "auditlog.action."+al.Func+"."+al.Action),
			al.Detail,
		}
		if err = cw.Write(cols); err != nil {
			c.Logger.Error(err)
			return
		}
	}
}
