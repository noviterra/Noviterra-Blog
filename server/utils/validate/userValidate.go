package validate

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

type userIdentityValidate interface {
	ValidateUserID(userID string) (string, error)
	ValidatePasswordFormat(password string) (string, error)
	ValidateVerificationCode(verificationCode string) (string, error)
	ValidateEmail(email string) (string, error)
	ValidateDoublePassword(newPassword,
		oldPassword string) (string, string, error)
}

type UserValidate struct {
	userIdentityValidate
}

func NewUserValidate(identityValidate userIdentityValidate) *UserValidate {
	return &UserValidate{
		identityValidate,
	}
}

const (
	minAge               = 0
	maxAge               = 100
	minNameLegth         = 2
	maxNameLength        = 20
	minInfomationLength  = 5
	maxInformationLength = 100
)

var (
	errAgeInvalid               = errors.New("年龄设置有误！")
	errNameEmpty                = errors.New("用户名为空！")
	errNameInvalid              = errors.New("用户名中出现了特殊字符！")
	errNameLegthInvalid         = errors.New("用户名长度不符合！")
	errInformationEmpty         = errors.New("个人简介为空！")
	errInformationLengthInvalid = errors.New("个人简介长度不符合！")
	errGenderInvalid            = errors.New("用户性别设置异常！")
)

type InputValidateUserInfo struct {
	Age              int
	Name             string
	Gender           int
	SelfIntroduction string
}

type OutputValidateUserInfo struct {
	Age              int
	Name             string
	Gender           int
	SelfIntroduction string
}

func (uv *UserValidate) ValidateUserInfo(in *InputValidateUserInfo) (*OutputValidateUserInfo, error) {
	age, err := uv.ValidateAge(in.Age)
	if err != nil {
		return nil, err
	}
	name, err := uv.ValidateUserName(in.Name)
	if err != nil {
		return nil, err
	}
	gender, err := uv.ValidateGender(in.Gender)
	if err != nil {
		return nil, err
	}
	selfIntroduction, err := uv.ValidateSelfIntroduction(in.SelfIntroduction)
	if err != nil {
		return nil, err
	}
	return &OutputValidateUserInfo{
		Age:              age,
		Name:             name,
		Gender:           gender,
		SelfIntroduction: selfIntroduction,
	}, nil
}

type InputChangePwdByEmail struct {
	ID               string
	Email            string
	VerificationCode string
	NewPassword      string
}

type OutputChangePwdByEmail struct {
	ID               string
	Email            string
	VerificationCode string
	NewPassword      string
}

func (uv *UserValidate) ValidateChangePwdByEmail(in InputChangePwdByEmail) (OutputChangePwdByEmail, error) {
	id, err := uv.ValidateUserID(in.ID)
	if err != nil {
		return OutputChangePwdByEmail{}, err
	}
	email, err := uv.ValidateEmail(in.Email)
	if err != nil {
		return OutputChangePwdByEmail{}, err
	}
	verificationCode, err := uv.ValidateVerificationCode(in.VerificationCode)
	if err != nil {
		return OutputChangePwdByEmail{}, err
	}
	newPassword, err := uv.ValidatePasswordFormat(in.NewPassword)
	if err != nil {
		return OutputChangePwdByEmail{}, err
	}

	return OutputChangePwdByEmail{
		ID:               id,
		Email:            email,
		VerificationCode: verificationCode,
		NewPassword:      newPassword,
	}, nil
}

type InputChangePwdByID struct {
	ID          string
	NewPassword string
	OldPassword string
}

type OutputChangePwdByID struct {
	ID          string
	NewPassword string
	OldPassword string
}

func (uv *UserValidate) ValidateChangePwdByID(in InputChangePwdByID) (OutputChangePwdByID, error) {
	id, err := uv.ValidateUserID(in.ID)
	if err != nil {
		return OutputChangePwdByID{}, err
	}
	newPassword, oldPassword, err := uv.ValidateDoublePassword(in.NewPassword, in.OldPassword)
	if err != nil {
		return OutputChangePwdByID{}, err
	}

	return OutputChangePwdByID{
		ID:          id,
		NewPassword: newPassword,
		OldPassword: oldPassword,
	}, nil
}

type InputBlackUser struct {
	ID         string
	OperatorID string
}

type OutputBlackUser struct {
	ID         string
	OperatorID string
}

func (uv *UserValidate) ValidateBlackUser(in InputBlackUser) (OutputBlackUser, error) {
	id, err := uv.ValidateUserID(in.ID)
	if err != nil {
		return OutputBlackUser{}, err
	}
	operator, err := uv.ValidateUserID(in.OperatorID)
	if err != nil {
		return OutputBlackUser{}, err
	}

	return OutputBlackUser{
		ID:         id,
		OperatorID: operator,
	}, nil
}

func (uv *UserValidate) ValidateAge(age int) (int, error) {
	// 限制年龄在0~100之间
	if age <= minAge || age >= maxAge {
		return -1, errAgeInvalid
	}

	return age, nil
}

var specialChar = regexp.MustCompile(`^[\p{Han}a-zA-Z·\-\s\.]+$`)

func (uv *UserValidate) ValidateUserName(name string) (string, error) {
	n := strings.TrimSpace(name)
	if n == "" {
		return "", errNameEmpty
	}

	// 姓名不能有特殊字符
	if !specialChar.MatchString(n) {
		return "", errNameInvalid
	}

	// 限制2~20字符内 一个中文占两个字符
	length := utf8.RuneCountInString(n)
	if length > maxNameLength || length < minNameLegth {
		return "", errNameLegthInvalid
	}

	return n, nil
}

func (uv *UserValidate) ValidateGender(gender int) (int, error) {
	if gender != 0 && gender != 1 && gender != 2 { // 分别对应未知 男 女
		return 0, errGenderInvalid
	}

	return gender, nil
}

func (uv *UserValidate) ValidateSelfIntroduction(selfIntroduction string) (string, error) {
	si := strings.TrimSpace(selfIntroduction)
	if si == "" {
		return "", errInformationEmpty
	}
	// 限制长度 100
	length := utf8.RuneCountInString(si)
	if length < minInfomationLength || length > maxInformationLength {
		return "", errInformationLengthInvalid
	}

	return si, nil
}
