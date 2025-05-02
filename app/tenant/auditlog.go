package tenant

import (
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xin"
)

func (tt *Tenant) AddAuditLog(tx sqlx.Sqlx, c *xin.Context, funact string, params ...string) error {
	uid, role := int64(0), models.RoleGuest

	au := GetAuthUser(c)
	if au != nil {
		uid = au.ID
		role = au.Role
	}

	cip := c.ClientIP()
	return tt.Schema.AddAuditLog(tx, uid, cip, role, funact, params...)
}
