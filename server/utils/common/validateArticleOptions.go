package common

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

type ValidateArticleInput struct {
	Title       string
	Information string
	Content     string
	Cover       string
	Tags        []string
}

type ValidateArticleOutput struct {
	Title       string
	Information string
	Content     string
	Cover       string
	Tags        []string
}

func (c *CommonUtils) ValidateArticle(in *ValidateArticleInput) (*ValidateArticleOutput, error) {
	title, err := c.CheckTitle(in.Title)
	if err != nil {
		return nil, err
	}

	information, err := c.CheckInformation(in.Information)
	if err != nil {
		return nil, err
	}

	content, err := c.CheckContent(in.Content)
	if err != nil {
		return nil, err
	}

	cover, err := c.CheckCover(in.Cover)
	if err != nil {
		return nil, err
	}

	tags, err := c.CheckTags(in.Tags)
	if err != nil {
		return nil, err
	}

	return &ValidateArticleOutput{
		Title:       title,
		Information: information,
		Content:     content,
		Cover:       cover,
		Tags:        tags,
	}, nil
}

var (
	TitleMaxLen       = 20
	InformationMaxLen = 200
	ContentMaxLen     = 50000
	CoverMaxLen       = 20
	TagMaxNumber      = 10
	TagMaxLen         = 10
)

var (
	ErrTitleEmpty                = errors.New("标题为空！")
	ErrTitleOutRange             = errors.New("标题超出限制字数！")
	ErrInformationEmpty          = errors.New("简介为空！")
	ErrInformationOutRange       = errors.New("简介超出字数限制！")
	ErrContentEmpty              = errors.New("文章正文为空！")
	ErrContentOutRange           = errors.New("文章正文超出字数限制！")
	ErrCoverEmpty                = errors.New("封面属性为空！")
	ErrCoverOutRange             = errors.New("封面属性超出了字数限制！")
	ErrorCoverHaveSpecialContent = errors.New("封面属性含义特殊字符！")
	ErrTagsOutNumberRange        = errors.New("标签个数超出限制！")
	ErrTagEmpty                  = errors.New("标签为空！")
	ErrTagOutRange               = errors.New("标签内容超出限制！")
	ErrTagDuplicated             = errors.New("标签重复出现！")
)

// CheckTitle 校验标题
func (c *CommonUtils) CheckTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", ErrTitleEmpty
	}
	if utf8.RuneCountInString(title) > TitleMaxLen {
		return "", ErrTitleOutRange
	}
	return title, nil
}

// CheckInformation 校验文章题记（主题介绍）
func (c *CommonUtils) CheckInformation(information string) (string, error) {
	information = strings.TrimSpace(information)
	if information == "" {
		return "", ErrInformationEmpty
	}
	if utf8.RuneCountInString(information) > InformationMaxLen {
		return "", ErrInformationOutRange
	}
	return information, nil
}

// CheckContent 校验文章正文
func (c *CommonUtils) CheckContent(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", ErrContentEmpty
	}
	if utf8.RuneCountInString(content) > ContentMaxLen {
		return "", ErrContentOutRange
	}
	return content, nil
}

// 只允许英文字母和数字，长度 1-128
var coverPattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// CheckCover 检查封面
func (c *CommonUtils) CheckCover(cover string) (string, error) {
	cover = strings.TrimSpace(cover)
	if cover == "" {
		return "", ErrCoverEmpty
	}
	if len(cover) > CoverMaxLen {
		return "", ErrCoverOutRange
	}
	if !coverPattern.MatchString(cover) {
		return "", ErrorCoverHaveSpecialContent
	}
	return cover, nil
}

// CheckTags 检查标签 返回去重、去空格后的字符串切片
func (c *CommonUtils) CheckTags(tags []string) ([]string, error) {
	if len(tags) <= 0 || len(tags) > TagMaxNumber {
		return nil, ErrTagsOutNumberRange
	}
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return nil, ErrTagEmpty
		}
		if utf8.RuneCountInString(tag) > TagMaxLen {
			return nil, ErrTagOutRange
		}
		if _, dup := seen[tag]; dup {
			return nil, ErrTagDuplicated
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result, nil
}
