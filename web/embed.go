package web

import (
	"embed"

	"github.com/askasoft/pango-assets/html/bootstrap5"
	"github.com/askasoft/pango-assets/html/bootswatch5"
	"github.com/askasoft/pango-assets/html/corejs"
	"github.com/askasoft/pango-assets/html/fontawesome6"
	"github.com/askasoft/pango-assets/html/jquery"
	"github.com/askasoft/pango-assets/html/plugins"
)

// Static embed static folder
var Statics = map[string]embed.FS{
	"bootstrap5":   bootstrap5.FS,
	"bootswatch5":  bootswatch5.FS,
	"fontawesome6": fontawesome6.FS,
	"corejs":       corejs.FS,
	"jquery":       jquery.FS,
	"plugins":      plugins.FS,
}

//go:embed assets favicon.ico
var FS embed.FS
