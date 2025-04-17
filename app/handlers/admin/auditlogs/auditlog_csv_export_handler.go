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
	alqa, err := bindAuditLogQueryArg(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	fm := tbsutil.GetAudioLogFuncMap(c.Locale)

	cw := csv.NewWriter(c.Writer)
	cw.UseCRLF = true
	defer cw.Flush()

	var cols []string
	err = tt.IterAuditLogs(app.SDB, alqa, func(al *models.AuditLogEx) error {
		if len(cols) == 0 {
			c.SetAttachmentHeader("auditlogs.csv")
			_, _ = c.Writer.WriteString(string(iox.BOM))

			cols = append(cols,
				tbs.GetText(c.Locale, "auditlog.id"),
				tbs.GetText(c.Locale, "auditlog.date"),
				tbs.GetText(c.Locale, "auditlog.user"),
				tbs.GetText(c.Locale, "auditlog.func"),
				tbs.GetText(c.Locale, "auditlog.action"),
				tbs.GetText(c.Locale, "auditlog.message"),
			)
			if err := cw.Write(cols); err != nil {
				return err
			}
		}

		if len(al.Params) > 0 {
			al.Detail = tbs.Format(c.Locale, "auditlog.detail."+al.Func+"."+al.Action, asg.Anys(al.Params)...)
		}

		cols = cols[:0]
		cols = append(cols,
			num.Ltoa(al.ID),
			app.FormatTime(al.Date),
			al.User,
			fm.SafeGet(al.Func, al.Func),
			tbs.Format(c.Locale, "auditlog.action."+al.Func+"."+al.Action),
			al.Detail,
		)
		return cw.Write(cols)
	})
	if err != nil {
		c.Logger.Error(err)
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}
}
