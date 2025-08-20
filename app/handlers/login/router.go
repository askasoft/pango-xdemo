package login

import (
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
)

func Router(rg *xin.RouterGroup) {
	rg.Use(app.XTP.Handle) // token protect
	rg.Use(app.XCN.Handle)

	rg.GET("/", Index)
	rg.POST("/login", Login)
	rg.POST("/mfa_enroll", LoginMFAEnroll)
	rg.GET("/logout", Logout)

	addLoginPasswordResetHandlers(rg.Group("/pwdrst"))
}

func addLoginPasswordResetHandlers(rg *xin.RouterGroup) {
	rg.GET("/", PasswordResetIndex)
	rg.POST("/send", PasswordResetSend)
	rg.GET("/reset/:token", PasswordResetConfirm)
	rg.POST("/reset/:token", PasswordResetExecute)
}
