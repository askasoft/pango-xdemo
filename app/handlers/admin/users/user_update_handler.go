package users

import (
	"errors"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/argutil"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango-xdemo/app/utils/pwdutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"gorm.io/gorm"
)

func UserNew(c *xin.Context) {
	user := &models.User{
		Role:   models.RoleViewer,
		Status: models.UserActive,
	}

	h := handlers.H(c)
	h["User"] = user
	userAddMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/user_detail_edit", h)
}

func userDetail(c *xin.Context, action string) {
	aid := num.Atol(c.Query("id"))
	if aid == 0 {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	user := &models.User{}
	err := app.GDB.Table(tt.TableUsers()).Where("id = ?", aid).Take(user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.AddError(err)
		c.JSON(http.StatusNotFound, handlers.E(c))
		return
	}
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	h := handlers.H(c)
	h["User"] = user

	userAddMaps(c, h)

	c.HTML(http.StatusOK, "admin/users/user_detail_"+action, h)
}

func UserView(c *xin.Context) {
	userDetail(c, "view")
}

func UserEdit(c *xin.Context) {
	userDetail(c, "edit")
}

func userValidateRole(c *xin.Context, role string) {
	if role != "" {
		au := tenant.AuthUser(c)
		urm := tbsutil.GetUserRoleMap(c.Locale, au.Role)
		if !urm.Contain(role) {
			c.AddError(vadutil.ErrInvalidField(c, "user.", "role"))
		}
	}
}

func userValidateStatus(c *xin.Context, status string) {
	if status != "" {
		sm := tbsutil.GetUserStatusMap(c.Locale)
		if !sm.Contain(status) {
			c.AddError(vadutil.ErrInvalidField(c, "user.", "status"))
		}
	}
}

func userValidatePassword(c *xin.Context, password string) {
	if password != "" {
		tt := tenant.FromCtx(c)

		if vs := tt.ValidatePassword(c.Locale, password); len(vs) > 0 {
			for _, v := range vs {
				c.AddError(&vadutil.ParamError{
					Param:   "password",
					Message: v,
				})
			}
		}
	}
}

func userBind(c *xin.Context) *models.User {
	user := &models.User{}
	if err := c.Bind(user); err != nil {
		vadutil.AddBindErrors(c, err, "user.")
	}

	userValidateRole(c, user.Role)
	userValidateStatus(c, user.Status)
	userValidatePassword(c, user.Password)
	return user
}

func UserCreate(c *xin.Context) {
	user := userBind(c)
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	user.ID = 0
	if user.Password == "" {
		user.Password = pwdutil.RandomPassword()
	}
	user.SetPassword(user.Password)
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	if err := app.GDB.Table(tt.TableUsers()).Create(user).Error; err != nil {
		if pgutil.IsUniqueViolation(err) {
			err = &vadutil.ParamError{
				Param:   "email",
				Message: tbs.Format(c.Locale, "user.error.duplicated", tbs.GetText(c.Locale, "user.email", "email"), user.Email),
			}
		}
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, xin.H{
		"user":    user,
		"success": tbs.GetText(c.Locale, "success.created"),
	})
}

func UserUpdate(c *xin.Context) {
	user := userBind(c)
	if user.ID == 0 {
		c.AddError(vadutil.ErrInvalidID(c))
	}
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	if user.Password == "" {
		eu := &models.User{}
		err := app.GDB.Table(tt.TableUsers()).Where("id = ?", user.ID).Take(eu).Error
		if err != nil {
			if pgutil.IsUniqueViolation(err) {
				err = &vadutil.ParamError{
					Param:   "email",
					Message: tbs.Format(c.Locale, "user.error.duplicated", tbs.GetText(c.Locale, "user.email", "email"), user.Email),
				}
			}
			c.AddError(err)
			c.JSON(http.StatusInternalServerError, handlers.E(c))
			return
		}

		// NOTE: we need reencrypt password, because password is encrypted by email
		user.SetPassword(eu.GetPassword())
	} else {
		user.SetPassword(user.Password)
	}

	user.UpdatedAt = time.Now()

	tx := app.GDB.Table(tt.TableUsers())
	tx = tx.Where("id = ?", user.ID)
	tx = tx.Where("role >= ?", au.Role)
	tx = tx.Select(
		"name",
		"email",
		"password",
		"role",
		"status",
		"cidr",
		"updated_at",
	)

	r := tx.Updates(user)
	if r.Error != nil {
		c.AddError(r.Error)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, xin.H{
		"user":    user,
		"success": tbs.Format(c.Locale, "user.success.updates", r.RowsAffected),
	})
}

type UserUpdatesArg struct {
	ID     string  `json:"id,omitempty" form:"id,strip"`
	Role   string  `json:"role,omitempty" form:"role,strip"`
	Status string  `json:"status,omitempty" form:"status,strip"`
	CIDR   *string `json:"cidr,omitempty" form:"cidr,strip" validate:"omitempty,cidrs"`
}

func (uua *UserUpdatesArg) IsEmpty() bool {
	return uua.Role == "" && uua.Status == "" && uua.CIDR == nil
}

func UserUpdates(c *xin.Context) {
	uua := &UserUpdatesArg{}
	if err := c.Bind(uua); err != nil {
		vadutil.AddBindErrors(c, err, "user.")
	}
	userValidateRole(c, uua.Role)
	userValidateStatus(c, uua.Status)

	if uua.IsEmpty() {
		c.AddError(errors.New(tbs.GetText(c.Locale, "error.request.invalid")))
	}
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	ids, ida := argutil.SplitIDs(uua.ID)
	if len(ids) == 0 && !ida {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	au := tenant.AuthUser(c)
	tt := tenant.FromCtx(c)

	tx := app.GDB.Table(tt.TableUsers())

	tx = tx.Where("id <> ?", au.ID)
	tx = tx.Where("role >= ?", au.Role)

	if len(ids) > 0 {
		tx = tx.Where("id IN ?", ids)
	}

	user := &models.User{}
	user.UpdatedAt = time.Now()

	cols := make([]string, 0, 4)
	cols = append(cols, "updated_at")

	if uua.Role != "" {
		user.Role = uua.Role
		cols = append(cols, "role")
	}
	if uua.Status != "" {
		user.Status = uua.Status
		cols = append(cols, "status")
	}
	if uua.CIDR != nil {
		user.CIDR = *uua.CIDR
		cols = append(cols, "cidr")
	}

	r := tx.Select(cols).Updates(user)
	if r.Error != nil {
		c.AddError(r.Error)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "user.success.updates", r.RowsAffected),
		"updates": uua,
	})
}

func UserDeletes(c *xin.Context) {
	ids, ida := argutil.SplitIDs(c.PostForm("id"))
	if len(ids) == 0 && !ida {
		c.AddError(vadutil.ErrInvalidID(c))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	au := tenant.AuthUser(c)
	tt := tenant.FromCtx(c)

	var cnt int64
	err := app.GDB.Transaction(func(db *gorm.DB) error {
		tx := db.Table(tt.TableUsers())
		tx = tx.Where("id <> ?", au.ID)
		tx = tx.Where("role >= ?", au.Role)
		if len(ids) > 0 {
			tx = tx.Where("id IN ?", ids)
		}
		r := tx.Delete(&models.User{})
		if r.Error != nil {
			return r.Error
		}
		cnt = r.RowsAffected

		return tt.ResetSequence(db, "users", models.UserStartID)
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "user.success.deletes", cnt),
	})
}

func UserDeleteBatch(c *xin.Context) {
	uq, err := userListArgs(c)
	if err != nil {
		vadutil.AddBindErrors(c, err, "user.")
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	if !uq.HasFilter() {
		c.AddError(errors.New(tbs.GetText(c.Locale, "error.param.nofilter")))
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)
	au := tenant.AuthUser(c)

	var cnt int64
	err = app.GDB.Transaction(func(db *gorm.DB) (err error) {
		tx := db.Table(tt.TableUsers())
		tx = tx.Where("id <> ?", au.ID)
		tx = uq.AddWhere(c, tx)
		r := tx.Delete(&models.User{})
		if err = r.Error; err != nil {
			return
		}
		cnt = r.RowsAffected

		return tt.ResetSequence(db, "users", models.UserStartID)
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "user.success.deletes", cnt),
	})
}
