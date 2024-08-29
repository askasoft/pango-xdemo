package otputil

import (
	"fmt"
	"testing"
	"time"
)

func TestTOTP(t *testing.T) {
	secret := "test@gmail.com"
	expire := time.Minute * 10

	totp := NewTOTP(secret, expire)

	layout := "2006-01-02 15:04:05"
	tm, _ := time.Parse(layout, "2020-01-01 10:05:00")
	passcode := totp.AtTime(tm)

	for i := 0; i < 10; i++ {
		td := time.Minute * time.Duration(i)
		t1 := tm.Add(td)
		t2 := t1.Add(-expire)

		fmt.Printf("+%s = %s\n", td, totp.AtTime(t1))

		if !totp.VerifyTime(passcode, t1) && !totp.VerifyTime(passcode, t2) {
			t.Errorf("%s: %s = %s, %s = %s", passcode, t1.Format(layout), totp.AtTime(t1), t2.Format(layout), totp.AtTime(t2))
		}
	}

}
