package cptutil

import (
	"fmt"
	"testing"
)

func TestEncryptPassword(t *testing.T) {
	username := "x@x.com"
	password := "trusttrusttrusttrust"

	encpass := MustEncrypt(username, password)

	decpass := MustDecrypt(username, encpass)
	if password != decpass {
		t.Errorf("%s: E(%s) != D(%s)", password, encpass, decpass)
	}

	fmt.Println(encpass)
}
