package otputil

import (
	"encoding/base32"
	"time"

	"github.com/askasoft/pango/str"
	"github.com/xlzd/gotp"
)

func NewSecret(s string) string {
	var encoder = base32.StdEncoding.WithPadding(base32.NoPadding)
	secret := encoder.EncodeToString(str.UnsafeBytes(s))
	return secret
}

func NewTOTP(secret string, interval time.Duration) *gotp.TOTP {
	secret = NewSecret(secret)
	totp := gotp.NewTOTP(secret, 6, int(interval.Seconds()), nil)
	return totp
}

func TOTPVerify(totp *gotp.TOTP, interval time.Duration, passcode string) bool {
	tm := time.Now()
	if totp.VerifyTime(passcode, tm) {
		return true
	}

	tm = tm.Add(-interval)
	return totp.VerifyTime(passcode, tm)
}
