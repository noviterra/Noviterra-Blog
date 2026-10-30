package validate

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

type NoticeValidate struct {
}

func NewNoticeValidate() *NoticeValidate {
	return &NoticeValidate{}
}

var (
	ErrTitleEmpty   = errors.New("公告标题输入为空！")
	ErrTitleTooLong = errors.New("公告标题长度超出了限制！")

	ErrMessageEmpty   = errors.New("公告内容输入为空！")
	ErrMessageTooLong = errors.New("公告内容长度超出了限制！")

	ErrEmailEmtpy   = errors.New("邮箱输入为空！")
	ErrEmailInvalid = errors.New("邮箱格式输入有误！")

	ErrLinkEmptys  = errors.New("链接输入为空！")
	ErrLinkInvalid = errors.New("链接格式输入有误！")

	ErrProcessEmpyt   = errors.New("处理状态不能为空！")
	ErrProcessInvalid = errors.New("不存在的处理状态！")
)

type InputValidateFriendLinkResult struct {
	Email   string
	Link    string
	Process string
}
type OutputValidateFriendLinkResult struct {
	Email   string
	Link    string
	Process string
}

func (nv *NoticeValidate) ValidateFriendLinkEmailResultInfo(in InputValidateFriendLinkResult) (OutputValidateFriendLinkResult, error) {
	link, err := nv.ValidateFriendLink(in.Link)
	if err != nil {
		return OutputValidateFriendLinkResult{}, err
	}

	email, err := nv.ValidateFriendEmail(in.Email)
	if err != nil {
		return OutputValidateFriendLinkResult{}, err
	}

	process, err := nv.ValidateFriendProcess(in.Process)
	if err != nil {
		return OutputValidateFriendLinkResult{}, err
	}

	return OutputValidateFriendLinkResult{
		Email:   email,
		Link:    link,
		Process: process,
	}, nil
}

var validProcesses = map[string]bool{
	"success": true,
	"failed":  true,
	"cancel":  true,
	"apply":   true,
}
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var linkRegex = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)

func (nv *NoticeValidate) ValidateFriendLink(link string) (string, error) {
	l := strings.TrimSpace(link)
	if l == "" {
		return "", ErrLinkEmptys
	}
	if !linkRegex.MatchString(l) {
		return "", ErrLinkInvalid
	}
	return l, nil
}

func (nv *NoticeValidate) ValidateFriendEmail(email string) (string, error) {
	e := strings.TrimSpace(email)
	if e == "" {
		return "", ErrEmailEmtpy
	}
	if !emailRegex.MatchString(e) {
		return "", ErrEmailInvalid
	}
	return e, nil
}

func (nv *NoticeValidate) ValidateFriendProcess(process string) (string, error) {
	p := strings.TrimSpace(process)
	if p == "" {
		return "", ErrProcessEmpyt
	}
	if !validProcesses[p] {
		return "", ErrProcessInvalid
	}
	return p, nil
}

type InputValidateNotice struct {
	Title   string
	Message string
}

type OutputValidateNotice struct {
	Title   string
	Message string
}

func (nv *NoticeValidate) ValidateNoticeInfo(in InputValidateNotice) (OutputValidateNotice, error) {
	title, err := nv.ValidateTitle(in.Title)
	if err != nil {
		return OutputValidateNotice{}, err
	}

	message, err := nv.ValidateMessage(in.Message)
	if err != nil {
		return OutputValidateNotice{}, err
	}

	return OutputValidateNotice{
		Title:   title,
		Message: message,
	}, nil
}

func (nv *NoticeValidate) ValidateTitle(title string) (string, error) {
	t := strings.TrimSpace(title)
	if t == "" {
		return "", ErrTitleEmpty
	}
	if utf8.RuneCountInString(t) > MaxNoticeTitleLength {
		return "", ErrTitleTooLong
	}
	return t, nil
}

func (nv *NoticeValidate) ValidateMessage(message string) (string, error) {
	m := strings.TrimSpace(message)
	if m == "" {
		return "", ErrMessageEmpty
	}
	if utf8.RuneCountInString(m) > MaxNoticeMessageLength {
		return "", ErrMessageTooLong
	}
	return m, nil
}
