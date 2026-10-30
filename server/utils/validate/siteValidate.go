package validate

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

type SiteValidate struct {
}

func NewSiteValidate() *SiteValidate {
	return &SiteValidate{}
}

const (
	MaxLinkLength          = 255
	MaxLinkNameLength      = 20
	MaxInformationLength   = 100
	MaxNoticeTitleLength   = 20
	MaxNoticeMessageLength = 200
)

var (
	ErrLinkEmpty          = errors.New("链接输入为空！")
	ErrLinkTooLong        = errors.New("链接长度超出了限制！")
	ErrLinkInvalidContent = errors.New("链接内容包含非法字符！")

	ErrLinkNameEmpty          = errors.New("网页名称输入为空！")
	ErrLinkNameTooLong        = errors.New("网页名称长度超出限制！")
	ErrLinkNameInvalidContent = errors.New("网页名称包含非法字符！")

	ErrInformationEmpty          = errors.New("网站简介输入为空！")
	ErrInformationTooLong        = errors.New("网站简介长度超出了限制！")
	ErrInformationInvalidContent = errors.New("网站简介包含非法字符！")
)

type InputValidateFriendLink struct {
	Link        string
	LinkName    string
	Information string
}
type OutputValidateFriendLink struct {
	Link        string
	LinkName    string
	Information string
}

func (sv *SiteValidate) ValidateFriendLinkInfo(in InputValidateFriendLink) (OutputValidateFriendLink, error) {
	link, err := sv.ValidateLink(in.Link)
	if err != nil {
		return OutputValidateFriendLink{}, err
	}

	linkName, err := sv.ValidateLinkName(in.LinkName)
	if err != nil {
		return OutputValidateFriendLink{}, err
	}

	information, err := sv.ValidateInformation(in.Information)
	if err != nil {
		return OutputValidateFriendLink{}, err
	}

	return OutputValidateFriendLink{
		Link:        link,
		LinkName:    linkName,
		Information: information,
	}, nil
}

func (sv *SiteValidate) ValidateLink(link string) (string, error) {
	l := strings.TrimSpace(link)
	if l == "" {
		return "", ErrLinkEmpty
	}
	if utf8.RuneCountInString(l) > MaxLinkLength {
		return "", ErrLinkTooLong
	}
	if sv.cotainsIllegalChar(link) {
		return "", ErrLinkInvalidContent
	}
	return l, nil
}

func (sv *SiteValidate) ValidateLinkName(linkName string) (string, error) {
	name := strings.TrimSpace(linkName)
	if name == "" {
		return "", ErrLinkNameEmpty
	}
	if utf8.RuneCountInString(name) > MaxLinkNameLength {
		return "", ErrLinkNameTooLong
	}
	if sv.cotainsIllegalChar(name) {
		return "", ErrLinkNameInvalidContent
	}
	return name, nil
}

func (sv *SiteValidate) ValidateInformation(information string) (string, error) {
	i := strings.TrimSpace(information)
	if i == "" {
		return "", ErrInformationEmpty
	}
	if utf8.RuneCountInString(i) > MaxInformationLength {
		return "", ErrInformationTooLong
	}
	if sv.cotainsIllegalChar(i) {
		return "", ErrInformationInvalidContent
	}
	return i, nil
}

// containsIllegalChar 拒绝控制字符与零宽字符
func (sv *SiteValidate) cotainsIllegalChar(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
		switch r {
		case '\u200b', '\u200c', '\u200d', '\ufeff': // 零宽空格/连接符/BOM
			return true
		}
	}
	return false
}
