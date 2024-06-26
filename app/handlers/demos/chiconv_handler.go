package demos

import (
	"net/http"

	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango/xin"
	"github.com/liuzl/gocc"
)

func ChiconvIndex(c *xin.Context) {
	h := handlers.H(c)

	c.HTML(http.StatusOK, "demos/chiconv", h)
}

var (
	cc_s2t, _ = gocc.New("s2t")
	cc_t2s, _ = gocc.New("t2s")
)

func chiconv(c *xin.Context, cc *gocc.OpenCC) {
	s := c.PostForm("s")
	t, err := cc.Convert(s)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{"success": t})
}

func ChiconvS2T(c *xin.Context) {
	chiconv(c, cc_s2t)
}

func ChiconvT2S(c *xin.Context) {
	chiconv(c, cc_t2s)
}
