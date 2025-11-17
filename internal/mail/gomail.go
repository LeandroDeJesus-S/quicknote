package mail

import (
	"gopkg.in/gomail.v2"
)

type goMailV2Mailer struct {
	conf *Config
}

// NewGOMailV2Mailer creates a new mailer using gomail v2.
// If config username and password are empty, the mailer won't use authentication.
func NewGOMailV2Mailer(cfg *Config) *goMailV2Mailer {
	return &goMailV2Mailer{conf: cfg}
}

func (m *goMailV2Mailer) Send(msg Message) error {
	if msg.From == "" {
		msg.From = m.conf.DefaultFrom
	}
	_msg := gomail.NewMessage()
	_msg.SetHeader("From", msg.From)
	_msg.SetHeader("To", msg.To...)
	_msg.SetHeader("Subject", msg.Subject)

	cttType := "text/plain"
	if msg.IsHTML {
		cttType = "text/html"
	}

	_msg.SetBody(cttType, string(msg.Body))

	if m.conf.Username != "" && m.conf.Password != "" {
		d := gomail.NewDialer(m.conf.Server, m.conf.Port, m.conf.Username, m.conf.Password)
		return d.DialAndSend(_msg)
	}

	d := gomail.Dialer{Host: m.conf.Server, Port: m.conf.Port}
	return d.DialAndSend(_msg)
}
