package main

import (
	"Noviterra/server/utils/email"
	"Noviterra/server/utils/snowflake"
	"fmt"
)

func testsSnowFlake() {
	sf, err := snowflake.NewSnowFlake(1)
	if err != nil {
		panic(err)
	}

	id := sf.GenerateSnowFlakeID()
	fmt.Printf("%d\n", id)
}

func testsEmailSender() {
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

func main() {
	testsEmailSender()
}
