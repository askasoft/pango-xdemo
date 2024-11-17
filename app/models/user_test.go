package models

import (
	"fmt"
	"testing"

	"github.com/askasoft/pango/str"
)

func TestEncryptPassword(t *testing.T) {
	u := &User{
		Email: "x@x.com",
	}

	for i := 1; i <= 128; i++ {
		u.SetPassword(str.Repeat("0", i))
		fmt.Printf("%d: %d\n", i, len(u.Password))
		if len(u.Password) > 200 {
			t.Errorf("%d: %d > 200", i, len(u.Password))
		}
	}
}
