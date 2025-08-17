package web

import (
	"embed"

	"github.com/askasoft/pangox-assets/html/bootstrap5"
	"github.com/askasoft/pangox-assets/html/bootswatch5/cosmo"
	"github.com/askasoft/pangox-assets/html/bootswatch5/flatly"
	"github.com/askasoft/pangox-assets/html/bootswatch5/pulse"
	"github.com/askasoft/pangox-assets/html/corejs"
	"github.com/askasoft/pangox-assets/html/docxjs"
	"github.com/askasoft/pangox-assets/html/fontawesome6"
	"github.com/askasoft/pangox-assets/html/jquery3"
	"github.com/askasoft/pangox-assets/html/jszip"
	"github.com/askasoft/pangox-assets/html/pdfjs"
	"github.com/askasoft/pangox-assets/html/pdfviewer"
	"github.com/askasoft/pangox-assets/html/plugins"
	"github.com/askasoft/pangox-assets/html/summernote"
)

// Static embed static folder
var Statics = map[string]embed.FS{
	"bootstrap5":         bootstrap5.FS,
	"bootswatch5/cosmo":  cosmo.FS,
	"bootswatch5/flatly": flatly.FS,
	"bootswatch5/pulse":  pulse.FS,
	"corejs":             corejs.FS,
	"fontawesome6":       fontawesome6.FS,
	"jquery3":            jquery3.FS,
	"jszip":              jszip.FS,
	"docxjs":             docxjs.FS,
	"pdfjs":              pdfjs.FS,
	"pdfviewer":          pdfviewer.FS,
	"plugins":            plugins.FS,
	"summernote":         summernote.FS,
}

//go:embed assets favicon.ico robots.txt
var FS embed.FS
