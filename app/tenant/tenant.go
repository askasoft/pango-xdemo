package tenant

import (
	"fmt"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango-xdemo/app/schema"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/xin"
)

func IsMultiTenant() bool {
	return schema.IsMultiTenant()
}

type Tenant struct {
	schema.Schema
	config map[string]string
}

func NewTenant(name string) *Tenant {
	tt := &Tenant{Schema: schema.Schema(name)}
	tt.config = tt.GetConfigMap()
	return tt
}

func GetSubdomain(c *xin.Context) (string, bool) {
	if !IsMultiTenant() {
		return "", true
	}

	host := c.Request.Host
	domain := app.Domain
	if host == domain {
		return "", true
	}

	suffix := "." + domain
	if !str.EndsWith(host, suffix) {
		return "", false
	}

	s := host[0 : len(host)-len(suffix)]
	return s, true
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
		return nil, fmt.Errorf("Invalid host %q", c.Request.Host)
	}

	if IsMultiTenant() {
		if s == "" {
			s = schema.DefaultSchema()
		}

		if ok, err := schema.CheckSchema(s); !ok || err != nil {
			return nil, err
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
