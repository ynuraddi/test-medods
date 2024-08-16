package smtp

import (
	"fmt"

	"github.com/go-mail/mail"
)

type Interface interface {
	SendMail(subject, body string, to ...string) error
	SendLoginFromNewIP(ip string, to string) error
}

var _ Interface = (*smtp)(nil)

type smtp struct {
	dialer *mail.Dialer
	from   string
}

func New(cfg *Config) *smtp {

	dialer := mail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	return &smtp{
		dialer: dialer,
		from:   cfg.From,
	}
}

func (s smtp) SendLoginFromNewIP(ip, to string) error {
	msg := fmt.Sprintf("Login from new IP address: %s", ip)
	return s.SendMail("Login from new IP.", msg, to)
}

func (s smtp) SendMail(subject, body string, to ...string) error {
	fmt.Printf("from[%s] to[%s] body[%s]\n", s.from, to[0], body)

	m := mail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	if err := s.dialer.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
