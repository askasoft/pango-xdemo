package tenant

import (
	"errors"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/utils/gormutil"
	"github.com/askasoft/pango-xdemo/app/utils/gormutil/mygorm"
	"github.com/askasoft/pango-xdemo/app/utils/gormutil/pggorm"
)

func DefaultSchema() string {
	return app.INI.GetString("database", "schema", "public")
}

func ExistsSchema(s string) (bool, error) {
	switch app.DBS["type"] {
	case "mysql":
		return mygorm.ExistsSchema(app.GDB, s)
	default:
		return pggorm.ExistsSchema(app.GDB, s)
	}
}

func ListSchemas() ([]string, error) {
	switch app.DBS["type"] {
	case "mysql":
		return mygorm.ListSchemas(app.GDB)
	default:
		return pggorm.ListSchemas(app.GDB)
	}
}

func CreateSchema(name string, comment string) error {
	switch app.DBS["type"] {
	case "mysql":
		return mygorm.CreateSchema(app.GDB, name)
	default:
		return pggorm.CreateSchema(app.GDB, name, comment)
	}
}

var ErrUnsupported = errors.New("unsupported")

func CommentSchema(name string, comment string) error {
	switch app.DBS["type"] {
	case "mysql":
		return ErrUnsupported
	default:
		return pggorm.CommentSchema(app.GDB, name, comment)
	}
}

func RenameSchema(old string, new string) error {
	switch app.DBS["type"] {
	case "mysql":
		return ErrUnsupported
	default:
		return pggorm.RenameSchema(app.GDB, old, new)
	}
}

func DeleteSchema(name string) error {
	switch app.DBS["type"] {
	case "mysql":
		return mygorm.DeleteSchema(app.GDB, name)
	default:
		return pggorm.DeleteSchema(app.GDB, name)
	}
}

func CountSchemas(sq *gormutil.SchemaQuery) (total int, err error) {
	switch app.DBS["type"] {
	case "mysql":
		return mygorm.CountSchemas(app.GDB, sq)
	default:
		return pggorm.CountSchemas(app.GDB, sq)
	}
}

func FindSchemas(sq *gormutil.SchemaQuery) (schemas []*gormutil.SchemaInfo, err error) {
	switch app.DBS["type"] {
	case "mysql":
		return mygorm.FindSchemas(app.GDB, sq)
	default:
		return pggorm.FindSchemas(app.GDB, sq)
	}
}
