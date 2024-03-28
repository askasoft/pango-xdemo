package admin

import (
	"errors"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"gorm.io/gorm"
)

func UserNew(c *xin.Context) {
	usr := &models.User{
		Role:   models.RoleViewer,
		Status: models.UserActive,
	}

	h := handlers.H(c)
	h["User"] = usr
	h["StatusMap"] = utils.GetUserStatusMap(c.Locale)
	h["RoleMap"] = utils.GetUserRoleMap(c.Locale)

	c.HTML(http.StatusOK, "admin/user_detail", h)
}

func UserDetail(c *xin.Context) {
	aid := num.Atol(c.Query("id"))
	if aid == 0 {
		c.AddError(utils.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	usr := &models.User{}
	r := app.DB.Table(tt.TableUsers()).Where("id = ?", aid).Take(usr)
	if errors.Is(r.Error, gorm.ErrRecordNotFound) {
		c.AddError(r.Error)
		c.JSON(http.StatusNotFound, handlers.E(c))
		return
	}
	if r.Error != nil {
		c.AddError(r.Error)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)
	h["User"] = usr
	h["StatusMap"] = utils.GetUserStatusMap(c.Locale)

	au := tenant.AuthUser(c)
	if au.IsSuper() {
		h["RoleMap"] = utils.GetSuperRoleMap(c.Locale)
	} else {
		h["RoleMap"] = utils.GetUserRoleMap(c.Locale)
	}

	c.HTML(http.StatusOK, "admin/user_detail", h)
}

func userBind(c *xin.Context) *models.User {
	usr := &models.User{}
	err := c.Bind(usr)
	if err != nil {
		utils.AddValidateErrors(c, err, "user.")
	}

	if !utils.ValidateCIDRs(usr.CIDR) {
		c.AddError(utils.ErrInvalidField(c, "user.", "cidr"))
	}

	sm := utils.GetUserStatusMap(c.Locale)
	if !sm.Contain(usr.Status) {
		c.AddError(utils.ErrInvalidField(c, "user.", "status"))
	}

	rm := utils.GetUserRoleMap(c.Locale)

	au := tenant.AuthUser(c)
	if au.IsSuper() {
		rm = utils.GetSuperRoleMap(c.Locale)
	}
	if !rm.Contain(usr.Role) {
		c.AddError(utils.ErrInvalidField(c, "user.", "role"))
	}

	return usr
}

func UserCreate(c *xin.Context) {
	usr := userBind(c)
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	usr.ID = 0
	if usr.Password == "" {
		usr.Password = str.RandLetterNumbers(16)
	}
	usr.SetPassword(usr.Password)
	usr.CreatedAt = time.Now()
	usr.UpdatedAt = usr.CreatedAt

	tt := tenant.FromCtx(c)
	err := app.DB.Table(tt.TableUsers()).Create(usr).Error
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	usr.Password = ""
	c.JSON(http.StatusOK, xin.H{
		"result":  usr,
		"success": tbs.GetText(c.Locale, "success.created"),
	})
}

func userUpdate(c *xin.Context, cols ...string) {
	usr := userBind(c)
	if usr.ID == 0 {
		c.AddError(utils.ErrInvalidID(c))
	}
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	if usr.Password == "" {
		eu := &models.User{}
		err := app.DB.Table(tt.TableUsers()).Where("id = ?", usr.ID).Take(eu).Error
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}

		// NOTE: we need reencrypt password, because password is encrypted by email
		usr.SetPassword(eu.GetPassword())
	} else {
		usr.SetPassword(usr.Password)
	}

	usr.UpdatedAt = time.Now()

	tx := app.DB.Table(tt.TableUsers())
	if len(cols) > 0 {
		tx = tx.Select(cols)
	}

	r := tx.Updates(usr)
	if r.Error != nil {
		c.AddError(r.Error)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	usr.Password = ""
	c.JSON(http.StatusOK, xin.H{
		"result":  usr,
		"success": tbs.GetText(c.Locale, "success.updated"),
	})
}

func UserUpdate(c *xin.Context) {
	userUpdate(c, "name", "email", "password", "role", "status", "cidr", "updated_at")
}

func UserDelete(c *xin.Context) {
	arg := &ArgIDs{}

	err := c.Bind(arg)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	cnt := len(arg.IDs)
	if cnt > 0 {
		au := tenant.AuthUser(c)
		tt := tenant.FromCtx(c)

		tx := app.DB.Table(tt.TableUsers())
		tx = tx.Where("role <> ? AND id <> ? AND id IN ?", models.RoleSuper, au.ID, arg.IDs)
		r := tx.Delete(&models.User{})

		err := r.Error
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}

		cnt = int(r.RowsAffected)
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "user.success.delete", cnt),
	})
}

func UserClear(c *xin.Context) {
	au := tenant.AuthUser(c)
	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.DB.Transaction(func(db *gorm.DB) error {
		r := db.Table(tt.TableUsers()).Where("role <> ? AND id <> ?", models.RoleSuper, au.ID).Delete(&models.User{})
		if r.Error != nil {
			return r.Error
		}
		cnt = r.RowsAffected

		return db.Exec(tt.ResetSequence("users", models.UserStartID)).Error
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "user.success.delete", cnt),
	})
}

func userStatusUpdate(c *xin.Context, enable bool) {
	arg := &ArgIDs{}

	err := c.Bind(arg)
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	cnt := len(arg.IDs)
	if cnt > 0 {
		au := tenant.AuthUser(c)
		tt := tenant.FromCtx(c)

		status := str.If(enable, models.UserActive, models.UserDisabled)
		tx := app.DB.Table(tt.TableUsers())
		tx = tx.Where("role <> ? AND id <> ? AND id IN ?", models.RoleSuper, au.ID, arg.IDs)
		r := tx.Update("status", status)

		err := r.Error
		if err != nil {
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}

		cnt = int(r.RowsAffected)
	}

	msg := str.If(enable, "user.success.enable", "user.success.disable")
	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, msg, cnt),
	})
}

func UserEnable(c *xin.Context) {
	userStatusUpdate(c, true)
}

func UserDisable(c *xin.Context) {
	userStatusUpdate(c, false)
}
