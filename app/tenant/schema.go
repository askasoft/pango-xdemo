package tenant

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/sqx/sqlx"
	"github.com/askasoft/pango/xsm"
	"github.com/askasoft/pango/xsm/pgsm/pggormsm"
	"github.com/askasoft/pango/xsm/pgsm/pgsqlxsm"
	"gorm.io/gorm"
)

func DefaultSchema() string {
	return app.INI.GetString("database", "schema", "public")
}

func GSM(db *gorm.DB) xsm.SchemaManager {
	return pggormsm.SM(db)
}

func SSM(db *sqlx.DB) xsm.SchemaManager {
	return pgsqlxsm.SM(db)
}

func SM() xsm.SchemaManager {
	if app.INI.GetString("internal", "xsm") == "sqlxsm" {
		return SSM(app.SDB)
	}
	return GSM(app.GDB)
}

func ExistsSchema(s string) (bool, error) {
	return SM().ExistsSchema(s)
}

func ListSchemas() ([]string, error) {
	return SM().ListSchemas()
}

func CreateSchema(name string, comment string) error {
	return SM().CreateSchema(name, comment)
}

func CommentSchema(name string, comment string) error {
	return SM().CommentSchema(name, comment)
}

func RenameSchema(old string, new string) error {
	return SM().RenameSchema(old, new)
}

func DeleteSchema(name string) error {
	return SM().DeleteSchema(name)
}

func CountSchemas(sq *xsm.SchemaQuery) (total int, err error) {
	return SM().CountSchemas(sq)
}

func FindSchemas(sq *xsm.SchemaQuery) (schemas []*xsm.SchemaInfo, err error) {
	return SM().FindSchemas(sq)
}
