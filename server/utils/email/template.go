package email

import "fmt"

func (es *EmailSender) TextEmailEditor(verificationCode, t string) (string, string) {
	switch t {
	case "login":
		return es.loginTextEmailEditor(verificationCode)
	case "register":
		return es.registerTextEmailEditor(verificationCode)
	case "forget_password":
		return es.forgetPasswordEditor(verificationCode)
	case "change_password":
		return es.changePasswordEditor(verificationCode)
	default:
		return "", ""
	}
}

func (es *EmailSender) registerTextEmailEditor(verificationCode string) (string, string) {
	subject := "NoviterraBlog注册验证码"
	body := fmt.Sprintf("【Noviterra】您的验证码为%s，有效期为5分钟，用于账号注册，泄露有风险，如非本人操作，请忽略。", verificationCode)

	return subject, body
}

func (es *EmailSender) loginTextEmailEditor(verificationCode string) (string, string) {
	subject := "NoviterraBlog登录验证码"
	body := fmt.Sprintf("【Noviterra】您的验证码为%s，有效期为5分钟，用于账号登录，泄露有风险，如非本人操作，请忽略。", verificationCode)

	return subject, body
}

func (es *EmailSender) forgetPasswordEditor(verificationCode string) (string, string) {
	subject := "NoviterraBlog密码找回验证码"
	body := fmt.Sprintf("【Noviterra】您的验证码为%s，有效期为5分钟，用于账号密码找回操作，泄露有风险，如非本人操作，请忽略。", verificationCode)

	return subject, body
}

func (es *EmailSender) changePasswordEditor(verificationCode string) (string, string) {
	subject := "NoviterraBlog密码修改验证码"
	body := fmt.Sprintf("【Noviterra】您的验证码为%s，有效期为5分钟，用于账号密码修改操作，泄露有风险，如非本人操作，请忽略。", verificationCode)

	return subject, body
}
