package smtputil

import (
	"crypto/tls"

	"github.com/askasoft/pango-xdemo/app"
	"github.com/askasoft/pango/net/email"
)

func SendHTMLMail(to string, subject, message string) error {
	sec := app.INI.Section("smtp")

	em := &email.Email{}
	if err := em.SetFrom(sec.GetString("fromaddr")); err != nil {
		return err
	}
	if err := em.AddTo(to); err != nil {
		return err
	}
	em.Subject = subject
	em.SetHTMLMsg(message)

	sender := &email.SMTPSender{
		Host:     sec.GetString("host", "localhost"),
		Port:     sec.GetInt("port", 25),
		Username: sec.GetString("username"),
		Password: sec.GetString("password"),
	}
	sender.Helo = "localhost"
	sender.Timeout = sec.GetDuration("timeout")
	if sec.GetBool("insecure") {
		sender.TLSConfig = &tls.Config{ServerName: sender.Host, InsecureSkipVerify: true} //nolint: gosec
	}

	if err := sender.Dial(); err != nil {
		return err
	}
	defer sender.Close()

	if err := sender.Login(); err != nil {
		return err
	}

	return sender.Send(em)
}
