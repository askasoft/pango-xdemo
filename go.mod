module github.com/askasoft/pango-xdemo

go 1.21

require (
	github.com/askasoft/pango v1.0.21
	github.com/askasoft/pango-assets v1.0.16
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.6.0
	github.com/liuzl/gocc v0.0.0-20231231122217-0372e1059ca5
	github.com/xlzd/gotp v0.1.0
)

require (
	github.com/adamzy/cedar-go v0.0.0-20170805034717-80a9c64b256d // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/liuzl/cedar-go v0.0.0-20170805034717-80a9c64b256d // indirect
	github.com/liuzl/da v0.0.0-20180704015230-14771aad5b1d // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
)

replace github.com/askasoft/pango => ../pango

replace github.com/askasoft/pango-assets => ../pango-assets
