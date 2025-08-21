package tenant

import (
	"errors"
	"fmt"
	"sync"

	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app"
	"github.com/askasoft/pangox-xdemo/app/schema"
)

type HostnameError struct {
	host string
}

func (he *HostnameError) Error() string {
	return fmt.Sprintf("Invalid host %q", he.host)
}

func IsHostnameError(err error) bool {
	var he *HostnameError
	return errors.As(err, &he)
}

type Tenant struct {
	schema.Schema
	config map[string]string
}

func NewTenant(name string) *Tenant {
	tt := &Tenant{Schema: schema.Schema(name)}
	tt.config = tt.getConfigMap()
	return tt
}

func IsMultiTenant() bool {
	return schema.IsMultiTenant()
}

func GetSubdomain(c *xin.Context) (string, bool) {
	if !IsMultiTenant() {
		return "", true
	}

	domain := c.RequestHostname()

	if domain == app.Domain() {
		return "", true
	}

	suffix := "." + app.Domain()
	if str.EndsWith(domain, suffix) {
		return domain[:len(domain)-len(suffix)], true
	}

	return "", false
}

const TENANT_CTXKEY = "TENANT"

func FromCtx(c *xin.Context) *Tenant {
	tt, ok := c.Get(TENANT_CTXKEY)
	if !ok {
		panic("Invalid Tenant!")
	}
	return tt.(*Tenant)
}

func FindAndSetTenant(c *xin.Context) (*Tenant, error) {
	if tt, ok := c.Get(TENANT_CTXKEY); ok {
		return tt.(*Tenant), nil
	}

	s, ok := GetSubdomain(c)
	if !ok {
		return nil, &HostnameError{c.Request.Host}
	}

	if s == "" {
		s = schema.DefaultSchema()
	}

	if IsMultiTenant() {
		ok, err := CheckSchema(s)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, &HostnameError{c.Request.Host}
		}
	}

	tt := NewTenant(s)
	c.Set(TENANT_CTXKEY, tt)
	return tt, nil
}

func Iterate(itf func(tt *Tenant) error) error {
	return schema.Iterate(func(sm schema.Schema) error {
		tt := NewTenant(string(sm))
		return itf(tt)
	})
}

func Create(name string, comment string) error {
	if err := schema.CreateSchema(name, comment); err != nil {
		return err
	}

	if err := schema.Schema(name).InitSchema(); err != nil {
		_ = schema.DeleteSchema(name)
		return err
	}

	return nil
}

// ---------------------------
var muSCMAS sync.Mutex

func CheckSchema(name string) (bool, error) {
	if v, ok := app.SCMAS.Get(name); ok {
		return v, nil
	}

	muSCMAS.Lock()
	defer muSCMAS.Unlock()

	// get again to prevent duplicated load
	if v, ok := app.SCMAS.Get(name); ok {
		return v, nil
	}

	exists, err := schema.ExistsSchema(name)
	if err != nil {
		return false, err
	}

	app.SCMAS.Set(name, exists)
	return exists, nil
}
