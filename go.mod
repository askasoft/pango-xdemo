module github.com/askasoft/pango-xdemo

go 1.21

toolchain go1.21.6

require (
	github.com/askasoft/pango v1.0.12
	github.com/askasoft/pango-assets v1.0.5
	github.com/google/uuid v1.6.0
	gorm.io/driver/mysql v1.5.1
	gorm.io/driver/postgres v1.5.7
	gorm.io/gorm v1.25.7
)

require (
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/askasoft/pango => ../pango

replace github.com/askasoft/pango-assets => ../pango-assets
