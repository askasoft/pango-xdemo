package user

import (
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/middles"
)

func Router(rg *xin.RouterGroup) {
	rg.Use(middles.AppAuth)      // app auth
	rg.Use(middles.IPProtect)    // IP protect
	rg.Use(middles.TokenProtect) // token protect

	addUserPwdchgHandlers(rg.Group("/pwdchg"))
}

func addUserPwdchgHandlers(rg *xin.RouterGroup) {
	rg.GET("/", PasswordChangeIndex)
	rg.POST("/change", PasswordChangeChange)
}
