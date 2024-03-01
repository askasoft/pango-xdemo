package utils

import (
	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/cpt"
	"github.com/askasoft/pango/log"
)

func Encrypt(s string) string {
	cryptor := cpt.NewAesCBC(app.Secret())

	es, err := cryptor.EncryptString(s)
	if err != nil {
		log.Error(err)
		return s
	}
	return es
}

func Decrypt(s string) string {
	cryptor := cpt.NewAesCBC(app.Secret())

	ds, err := cryptor.DecryptString(s)
	if err != nil {
		log.Error(err)
		return s
	}
	return ds
}
