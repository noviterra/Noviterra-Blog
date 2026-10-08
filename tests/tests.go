package main

import (
	"Noviterra/server/utils/email"
	"fmt"
)

func main() {
	emailSender := email.NewEmailSender(email.EmailConfig{
		Host:     "smtp.163.com",
		Port:     465,
		User:     "noviterra@163.com",
		From:     "noviterra@163.com",
		Password: "JPPYLEdiL6TAQyHs",
	})

	err := emailSender.SendTextEmails([]string{"3489673482@qq.com"}, "Tests", "This is a test's email.")
	if err != nil {
		fmt.Println(err)
	}
}
