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
	"github.com/askasoft/pango/cog"
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

	usr := &models.User{}
	r := app.GDB.Table(tt.TableUsers()).Where("id = ?", aid).Take(usr)
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
		var rm *cog.LinkedHashMap[string, string]

		au := tenant.AuthUser(c)
		if au.IsSuper() {
			rm = tbsutil.GetSuperRoleMap(c.Locale)
		} else {
			rm = tbsutil.GetUserRoleMap(c.Locale)
		}
		if !rm.Contain(role) {
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
	usr := &models.User{}
	if err := c.Bind(usr); err != nil {
		vadutil.AddBindErrors(c, err, "user.")
	}

	userValidateCIDR(c, usr.CIDR)
	userValidateRole(c, usr.Role)
	userValidateStatus(c, usr.Status)

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
	if err := app.GDB.Table(tt.TableUsers()).Create(usr).Error; err != nil {
		if pgutil.IsUniqueViolation(err) {
			err = &vadutil.ParamError{
				Param:   "email",
				Message: tbs.Format(c.Locale, "user.error.email.dup", tbs.GetText(c.Locale, "user.email", "email"), usr.Email),
			}
		}
		c.AddError(err)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	usr.Password = ""
	c.JSON(http.StatusOK, xin.H{
		"user":    usr,
		"success": tbs.GetText(c.Locale, "success.created"),
	})
}

var userUpdatables = []string{
	"name",
	"email",
	"password",
	"role",
	"status",
	"cidr",
	"updated_at",
}

func UserUpdate(c *xin.Context) {
	usr := userBind(c)
	if usr.ID == 0 {
		c.AddError(vadutil.ErrInvalidID(c))
	}
	if len(c.Errors) > 0 {
		c.JSON(http.StatusBadRequest, handlers.E(c))
		return
	}

	tt := tenant.FromCtx(c)

	if usr.Password == "" {
		eu := &models.User{}
		err := app.GDB.Table(tt.TableUsers()).Where("id = ?", usr.ID).Take(eu).Error
		if err != nil {
			if pgutil.IsUniqueViolation(err) {
				err = &vadutil.ParamError{
					Param:   "email",
					Message: tbs.Format(c.Locale, "user.error.email.dup", tbs.GetText(c.Locale, "user.email", "email"), usr.Email),
				}
			}
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

	r := app.GDB.Table(tt.TableUsers()).Select(userUpdatables).Updates(usr)
	if r.Error != nil {
		c.AddError(r.Error)
		c.JSON(http.StatusInternalServerError, handlers.E(c))
		return
	}

	usr.Password = ""
	c.JSON(http.StatusOK, xin.H{
		"user":    usr,
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

	ids := handlers.SplitIDs(uua.ID)
	if uua.ID != "*" && len(ids) == 0 {
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
		if !au.IsSuper() {
			tx = tx.Where("role <> ?", models.RoleSuper)
		}

		if uua.ID != "*" {
			tx = tx.Where("id IN ?", ids)
		}

		usr := &models.User{}

		usr.UpdatedAt = time.Now()

		cols := make([]string, 0, 8)
		cols = append(cols, "updated_at")

		if uua.Role != "" {
			usr.Role = uua.Role
			cols = append(cols, "role")
		}
		if uua.Status != "" {
			usr.Status = uua.Status
			cols = append(cols, "status")
		}
		if uua.Ucidr {
			usr.CIDR = uua.CIDR
			cols = append(cols, "cidr")
		}

		r := tx.Select(cols).Updates(usr)
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
	id := c.PostForm("id")
	ids := handlers.SplitIDs(id)

	if id != "*" && len(ids) == 0 {
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
		if !au.IsSuper() {
			tx = tx.Where("role <> ?", models.RoleSuper)
		}
		if id != "*" {
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
