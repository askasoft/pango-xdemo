package api

import (
	"net/http"

	"github.com/askasoft/pangox/xin"
)

func MyIP(c *xin.Context) {
	c.String(http.StatusOK, c.ClientIP())
}

func MyHeader(c *xin.Context) {
	c.IndentedJSON(http.StatusOK, c.Request.Header)
}

func MyGet(c *xin.Context) {
	c.IndentedJSON(http.StatusOK, c.Querys())
}

func MyPost(c *xin.Context) {
	c.IndentedJSON(http.StatusOK, c.PostForms())
}
