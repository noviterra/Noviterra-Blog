package email

import "fmt"

func (es *EmailSender) TextFriendLinkEmail(link, t string) (string, string) {
	// 友链结果邮件通知：申请后发送一条、申请成功后发送一条、申请失败发送一条、被下架发送一条
	switch t {
	case "apply":
		return es.TextApplyFriendLinkMessage(link)
	case "success":
		return es.TextApplyFriendLinkSuccessMessage(link)
	case "fail":
		return es.TextApplyFriendLinkFailMessage(link)
	case "cancle":
		return es.TextFriendLinkIsCancledMessage(link)
	default:
		return "", ""
	}
}

func (es *EmailSender) TextApplyFriendLinkMessage(link string) (string, string) {
	subject := "Noviterra友链申请进度"
	body := fmt.Sprintf("【Noviterra】感谢您对Noviterra的信任与支持！您的友链申请【%s】 已收到，我们将加急审核您的网站内容。"+
		"审核完成后，我们会在第一时间通过邮件告知您审核结果，请您耐心等待。如有疑问，欢迎随时联系我们。", link)

	return subject, body
}

func (es *EmailSender) TextApplyFriendLinkSuccessMessage(link string) (string, string) {
	subject := "Noviterra友链申请通过"
	body := fmt.Sprintf("【Noviterra】恭喜！您的友链申请已审核通过，您的网站链接【%s】 已成功展示在Noviterra友链页面中。"+
		"感谢您的支持，期待双方共同成长，如有任何变动请及时联系我们。", link)

	return subject, body
}

func (es *EmailSender) TextApplyFriendLinkFailMessage(link string) (string, string) {
	subject := "Noviterra友链审核结果"
	body := fmt.Sprintf("【Noviterra】感谢您申请Noviterra友链。经审核，您的网站【%s】 暂不符合Noviterra友链收录标准，本次申请未能通过。"+
		"如有疑问，欢迎回复邮件与我们进一步沟通。", link)

	return subject, body
}

func (es *EmailSender) TextFriendLinkIsCancledMessage(link string) (string, string) {
	subject := "Noviterra友链状态变更"
	body := fmt.Sprintf("【Noviterra】您的友链已被下架。可能原因为：您的网站【%s】 长时间无法访问、友链接失效或不再符合Noviterra友链收录标准。"+
		"如需重新申请，欢迎再次提交友链申请。", link)

	return subject, body
}
