package web

import (
	"embed"

	"github.com/askasoft/pango-assets/html/bootstrap3"
	"github.com/askasoft/pango-assets/html/corejs"
	"github.com/askasoft/pango-assets/html/fontawesome4"
	"github.com/askasoft/pango-assets/html/jquery"
	"github.com/askasoft/pango-assets/html/plugins"
)

// Favicon embed favicon.ico
//
//go:embed favicon.ico
var Favicon []byte

// Static embed static folder
var Statics = map[string]embed.FS{
	"bootstrap3":   bootstrap3.FS,
	"fontawesome4": fontawesome4.FS,
	"corejs":       corejs.FS,
	"jquery":       jquery.FS,
	"plugins":      plugins.FS,
}

// Assets embed assets folder
//
//go:embed assets
var Assets embed.FS
