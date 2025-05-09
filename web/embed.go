package web

import (
	"embed"

	"github.com/askasoft/pango-assets/html/bootstrap5"
	"github.com/askasoft/pango-assets/html/bootswatch5/cosmo"
	"github.com/askasoft/pango-assets/html/bootswatch5/flatly"
	"github.com/askasoft/pango-assets/html/bootswatch5/pulse"
	"github.com/askasoft/pango-assets/html/corejs"
	"github.com/askasoft/pango-assets/html/fontawesome6"
	"github.com/askasoft/pango-assets/html/jquery3"
	"github.com/askasoft/pango-assets/html/plugins"
	"github.com/askasoft/pango-assets/html/summernote"
)

// Static embed static folder
var Statics = map[string]embed.FS{
	"bootstrap5":         bootstrap5.FS,
	"bootswatch5/cosmo":  cosmo.FS,
	"bootswatch5/flatly": flatly.FS,
	"bootswatch5/pulse":  pulse.FS,
	"fontawesome6":       fontawesome6.FS,
	"corejs":             corejs.FS,
	"jquery3":            jquery3.FS,
	"plugins":            plugins.FS,
	"summernote":         summernote.FS,
}

//go:embed assets favicon.ico robots.txt
var FS embed.FS
