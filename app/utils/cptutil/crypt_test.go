package cptutil

import (
	"fmt"
	"testing"
)

func TestEncryptPassword(t *testing.T) {
	username := "x@x.com"
	password := "trusttrusttrusttrust"

	encpass := Encrypt(username, password)

	decpass := Decrypt(username, encpass)
	if password != decpass {
		t.Errorf("%s: E(%s) != D(%s)", password, encpass, decpass)
	}

	fmt.Println(encpass)
}
