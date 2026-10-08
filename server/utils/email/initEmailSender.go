package email

import "gopkg.in/gomail.v2"

// 邮件配置结构体 依赖注入时接收参数
type EmailConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

type EmailSender struct {
	Dialer *gomail.Dialer
	From   string
}

func NewEmailSender(ec EmailConfig) *EmailSender {
	dialer := gomail.NewDialer(ec.Host, ec.Port, ec.User, ec.Password)
	dialer.SSL = true
	return &EmailSender{
		Dialer: dialer,
		From:   ec.From,
	}
}
