package cptutil

import (
	"crypto/sha256"
	"fmt"

	"github.com/askasoft/pango/cpt"
	"github.com/askasoft/pango/str"
)

func Encrypt(secret, s string) string {
	cryptor := cpt.NewAesCBC(secret)

	es, err := cryptor.EncryptString(s)
	if err != nil {
		panic(err)
	}
	return es
}

func Decrypt(secret, s string) string {
	cryptor := cpt.NewAesCBC(secret)

	ds, err := cryptor.DecryptString(s)
	if err != nil {
		panic(err)
	}
	return ds
}

func Hash(s string) string {
	h := sha256.New()
	h.Write(str.UnsafeBytes(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
