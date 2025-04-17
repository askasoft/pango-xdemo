package user

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/middles"
	"github.com/askasoft/pango/xin"
)

func Router(rg *xin.RouterGroup) {
	rg.Use(middles.AppAuth)   // app auth
	rg.Use(middles.IPProtect) // IP protect
	rg.Use(app.XTP.Handle)    // token protect

	addUserPwdchgHandlers(rg.Group("/pwdchg"))
}

func addUserPwdchgHandlers(rg *xin.RouterGroup) {
	rg.GET("/", PasswordChangeIndex)
	rg.POST("/change", PasswordChangeChange)
}
