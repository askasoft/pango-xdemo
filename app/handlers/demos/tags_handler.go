package demos

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango/cog"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

type tagsArg struct {
	Text     string   `form:"text"`
	Checks   []string `form:"checks"`
	Radios   string   `form:"radios"`
	Select   string   `form:"select"`
	Textarea string   `form:"textarea"`
}

func TagsIndex(c *xin.Context) {
	h := handlers.H(c)

	a := &tagsArg{}
	_ = c.Bind(a)

	checks := &cog.LinkedHashMap[string, string]{}
	_ = checks.UnmarshalJSON(str.UnsafeBytes(tbs.GetText(c.Locale, "demo.tags.checks")))

	radios := &cog.LinkedHashMap[string, string]{}
	_ = radios.UnmarshalJSON(str.UnsafeBytes(tbs.GetText(c.Locale, "demo.tags.radios")))

	selects := &cog.LinkedHashMap[string, string]{}
	_ = selects.UnmarshalJSON(str.UnsafeBytes(tbs.GetText(c.Locale, "demo.tags.selects")))

	h["ChecksList"] = checks
	h["RadiosList"] = radios
	h["SelectList"] = selects
	h["Arg"] = a

	c.HTML(http.StatusOK, "demos/tags", h)
}
