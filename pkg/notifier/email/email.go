package email

import (
	"context"
	"fmt"
	"net/smtp"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
}

type Email struct {
	cfg  Config
	auth smtp.Auth
}

func New(cfg Config) *Email {
	return &Email{
		cfg:  cfg,
		auth: smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host),
	}
}

func (e *Email) Notify(ctx context.Context, message string) error {
	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)
	return smtp.SendMail(addr, e.auth, e.cfg.From, e.cfg.To, []byte(message))
}
