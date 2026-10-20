package main

import (
	"Noviterra/server/utils/bloomfilter"
	"Noviterra/server/utils/common"
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

func testsBloomFilter() {
	userBloomFilter, err := bloomfilter.NewUserBloomFilter(100, 0.001)
	if err != nil {
		panic(err)
	}
	data := []string{"anyone", "syy663", "swh5511", "noviterra"}
	for _, d := range data {
		userBloomFilter.AddUsers(d)
	}

	// 测试
	fmt.Println(userBloomFilter.IsExistUser("anyone"))
	fmt.Println(userBloomFilter.IsExistUser("HelloWorld!"))
	fmt.Println(userBloomFilter.IsExistUser("BadRequest"))
	fmt.Println(userBloomFilter.IsExistUser("novitene"))
	fmt.Println(userBloomFilter.IsExistUser("swh5511"))
}

func testsCommonUtils() {
	// 235729194842062848 真实雪花ID
	commonUtils := common.NewCommonUtils()

	ok, err := commonUtils.CheckArticleID("235729194842062848")
	if err != nil {
		panic(err)
	} else if !ok {
		fmt.Println("NO")
	} else {
		fmt.Println("YES")
	}

}

func main() {
	testsSnowFlake()
}
