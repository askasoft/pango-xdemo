module github.com/askasoft/pango-xdemo

go 1.20

require (
	github.com/askasoft/pango v1.0.11
	github.com/askasoft/pango-assets v1.0.4
	gorm.io/driver/mysql v1.5.1
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)

require (
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.13.0 // indirect
)

replace github.com/askasoft/pango => ../pango

replace github.com/askasoft/pango-assets => ../pango-assets
