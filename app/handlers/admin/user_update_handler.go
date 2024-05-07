package admin

import (
	"errors"
	"net/http"
	"time"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/handlers"
	"github.com/askasoft/pango-xdemo/app/models"
	"github.com/askasoft/pango-xdemo/app/tenant"
	"github.com/askasoft/pango-xdemo/app/utils/pgutil"
	"github.com/askasoft/pango-xdemo/app/utils/tbsutil"
	"github.com/askasoft/pango-xdemo/app/utils/vadutil"
	"github.com/askasoft/pango/num"
	"github.com/askasoft/pango/str"
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

	c.HTML(http.StatusOK, "admin/user_detail", h)
}

func UserDetail(c *xin.Context) {
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

	c.HTML(http.StatusOK, "admin/user_detail", h)
}

func userValidateCIDR(c *xin.Context, cidr string) {
	if !vadutil.ValidateCIDRs(cidr) {
		c.AddError(vadutil.ErrInvalidField(c, "user.", "cidr"))
	}
}

func userValidateRole(c *xin.Context, role string) {
	if role != "" {
		urm := tenant.GetUserRoleMap(c)
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

func userBind(c *xin.Context) *models.User {
	user := &models.User{}
	if err := c.Bind(user); err != nil {
		vadutil.AddBindErrors(c, err, "user.")
	}

	userValidateCIDR(c, user.CIDR)
	userValidateRole(c, user.Role)
	userValidateStatus(c, user.Status)

	return user
}

func UserCreate(c *xin.Context) {
	user := userBind(c)
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	user.ID = 0
	if user.Password == "" {
		user.Password = str.RandLetterNumbers(16)
	}
	user.SetPassword(user.Password)
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	tt := tenant.FromCtx(c)
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
	ID     string `json:"id,omitempty" form:"id,strip"`
	Role   string `json:"role" form:"role,strip"`
	Status string `json:"status" form:"status,strip"`
	CIDR   string `json:"cidr" form:"cidr,strip"`
	Ucidr  bool   `json:"ucidr" form:"ucidr"`
}

func UserUpdates(c *xin.Context) {
	uua := &UserUpdatesArg{}
	if err := c.Bind(uua); err != nil {
		vadutil.AddBindErrors(c, err, "user.")
	}
	userValidateCIDR(c, uua.CIDR)
	userValidateRole(c, uua.Role)
	userValidateStatus(c, uua.Status)

	if uua.Role == "" && uua.Status == "" && !uua.Ucidr {
		c.AddError(errors.New(tbs.GetText(c.Locale, "error.request.invalid")))
	}

	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	ids, ida := handlers.SplitIDs(uua.ID)
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
		if uua.Ucidr {
			user.CIDR = uua.CIDR
			cols = append(cols, "cidr")
		}

		r := tx.Select(cols).Updates(user)
		cnt = r.RowsAffected
		return r.Error
	})
	if err != nil {
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	c.JSON(http.StatusOK, xin.H{
		"success": tbs.Format(c.Locale, "user.success.updates", cnt),
		"updates": uua,
	})
}

func UserDeletes(c *xin.Context) {
	ids, ida := handlers.SplitIDs(c.PostForm("id"))
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

		return db.Exec(tt.ResetSequence("users", models.UserStartID)).Error
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
