package email

import "gopkg.in/gomail.v2"

// 发送文本文件
func (es *EmailSender) SendTextEmails(to []string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", es.From)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	return es.Dialer.DialAndSend(m)
}

// 发送HTML邮件
func (es *EmailSender) SendHTMLEmail(to []string, subject, htmlBody string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", es.From)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	return es.Dialer.DialAndSend(m)
}

// 带附件邮件
func (es *EmailSender) SendEmailWithFile(to, attachments []string, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", es.From)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	for _, file := range attachments {
		m.Attach(file)
	}

	return es.Dialer.DialAndSend(m)
}
