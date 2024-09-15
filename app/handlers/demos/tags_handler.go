package demos

import (
	"errors"
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
)

type tagsArg struct {
	Text     string   `form:"text"`
	Hchecks  []string `form:"hchecks"`
	Vchecks  []string `form:"vchecks"`
	Ochecks  []string `form:"ochecks"`
	Hradios  string   `form:"hradios"`
	Vradios  string   `form:"vradios"`
	Fselect  string   `form:"fselect"`
	Nselect  string   `form:"nselect"`
	Mselect  []string `form:"mselect"`
	Textarea string   `form:"textarea"`
	Htmledit string   `form:"htmledit"`
}

func TagsIndex(c *xin.Context) {
	h := handlers.H(c)

	a := &tagsArg{
		Ochecks:  []string{"c2"},
		Htmledit: "<pre>HTML本文</pre>",
	}
	_ = c.Bind(a)

	checks := &linkedhashmap.LinkedHashMap[string, string]{}
	_ = checks.UnmarshalJSON(str.UnsafeBytes(tbs.GetText(c.Locale, "demos.tags.checks")))

	radios := &linkedhashmap.LinkedHashMap[string, string]{}
	_ = radios.UnmarshalJSON(str.UnsafeBytes(tbs.GetText(c.Locale, "demos.tags.radios")))

	selects := &linkedhashmap.LinkedHashMap[string, string]{}
	_ = selects.UnmarshalJSON(str.UnsafeBytes(tbs.GetText(c.Locale, "demos.tags.selects")))

	h["ChecksList"] = checks
	h["RadiosList"] = radios
	h["SelectList"] = selects
	h["Arg"] = a

	c.AddError(errors.New(str.Repeat("Error message. ", 20)))
	h["Warning"] = str.Repeat("Warning message. ", 20)
	h["Warnings"] = []string{
		"Warning message 1.",
		"Warning message 2.",
	}
	h["Message"] = str.Repeat("Information message. ", 20)
	h["Messages"] = []string{
		"Information message 1.",
		"Information message 2.",
	}
	h["Success"] = str.Repeat("Success message. ", 20)
	h["Successes"] = []string{
		"Success message 1.",
		"Success message 2.",
	}

	c.HTML(http.StatusOK, "demos/tags", h)
}
