package tenant

import (
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xin"
)

func (tt *Tenant) AddAuditLog(tx sqlx.Sqlx, c *xin.Context, funact string, params ...string) error {
	var uid int64
	au := GetAuthUser(c)
	if au != nil {
		uid = au.ID
	}

	cip := c.ClientIP()
	return tt.Schema.AddAuditLog(tx, uid, cip, funact, params...)
}
