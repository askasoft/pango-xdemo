package login

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/xin"
)

func Router(rg *xin.RouterGroup) {
	rg.Use(app.XTP.Handler()) // token protect
	rg.Use(app.XCN.Handler())

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
