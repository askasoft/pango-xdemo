module github.com/askasoft/pango-xdemo/cmd

go 1.23.0

require (
	github.com/askasoft/pango v1.0.21
	github.com/askasoft/pango-xdemo v0.0.0-00010101000000-000000000000
	gorm.io/driver/mysql v1.5.7
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.30.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.5 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)

replace github.com/askasoft/pango-xdemo => ../

replace github.com/askasoft/pango => ../../pango
