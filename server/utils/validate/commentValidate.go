package validate

import (
	"errors"
	"strings"
)

type commentIdentityValidate interface {
	ValidateUserID(userID string) (string, error)
}

type commentArticleValidate interface {
	ValidateSnowID(id int64) (int64, error)
	ValidatePage(page, pageSize int32) (int, int, error)
}

type CommentValidate struct {
	commentIdentityValidate
	commentArticleValidate
}

func NewCommentValidate(identityValidate commentIdentityValidate,
	articleValidate commentArticleValidate) *CommentValidate {
	return &CommentValidate{
		identityValidate,
		articleValidate,
	}
}

var (
	ErrActionEmpty   = errors.New("点赞操作内容为空！")
	ErrActionInvalid = errors.New("点赞操作类型有误！")

	ErrParentIDInvalid = errors.New("评论ID格式错误！")
	ErrAddressEmpty    = errors.New("评论地址为空！")
	ErrAddressInvalid  = errors.New("评论地址格式错误！")
	ErrContentEmpty    = errors.New("评价内容为空！")

	ErrFieldEmpty   = errors.New("查询类型为空！")
	ErrFieldInvalid = errors.New("查询类型格式错误！")
)

type InputLikeComment struct {
	ArticleID int64
	UserID    string
	Action    string
}

type OutputLikeComment struct {
	ArticleID int64
	UserID    string
	Action    string
}

func (cv *CommentValidate) ValidateLikeComment(in InputLikeComment, field string) (OutputLikeComment, error) {
	articleID, err := cv.ValidateSnowID(in.ArticleID)
	if err != nil {
		return OutputLikeComment{}, err
	}

	userID, err := cv.ValidateUserID(in.UserID)
	if err != nil {
		return OutputLikeComment{}, err
	}

	action, err := cv.ValidateAction(in.Action, field)
	if err != nil {
		return OutputLikeComment{}, err
	}

	return OutputLikeComment{
		ArticleID: articleID,
		UserID:    userID,
		Action:    action,
	}, nil
}

func (cv *CommentValidate) ValidateAction(action, filed string) (string, error) {
	a := strings.TrimSpace(action)
	if a == "" {
		return "", ErrActionEmpty
	}
	if a != filed && a != "un"+filed {
		return "", ErrActionInvalid
	}
	return a, nil
}

type InputComment struct {
	Address   int64
	ParentID  int64
	SpeakerID string
	Content   string
}

type OutputComment struct {
	Address   int64
	ParentID  int64
	SpeakerID string
	Content   string
}

func (cv *CommentValidate) ValidateComment(in InputComment) (OutputComment, error) {
	parentID, err := cv.ValidateCommentParentID(in.ParentID)
	if err != nil {
		return OutputComment{}, err
	}

	speakerID, err := cv.ValidateUserID(in.SpeakerID)
	if err != nil {
		return OutputComment{}, err
	}

	address, err := cv.ValidateCommentAddress(in.Address)
	if err != nil {
		return OutputComment{}, err
	}

	content, err := cv.ValidateCommentContent(in.Content)
	if err != nil {
		return OutputComment{}, err
	}

	return OutputComment{
		Address:   address,
		ParentID:  parentID,
		SpeakerID: speakerID,
		Content:   content,
	}, nil
}

type InputUserLikeCollectArticleList struct {
	UserID   string
	Page     int32
	PageSize int32
	Fileds   string
}

type OutputUserLikeCollectArticleList struct {
	UserID   string
	Page     int
	PageSize int
	Fileds   string
}

func (cv *CommentValidate) ValidateUserLikeCollectArticleList(in InputUserLikeCollectArticleList) (OutputUserLikeCollectArticleList, error) {
	userID, err := cv.ValidateUserID(in.UserID)
	if err != nil {
		return OutputUserLikeCollectArticleList{}, err
	}

	page, pagesize, err := cv.ValidatePage(in.Page, in.PageSize)
	if err != nil {
		return OutputUserLikeCollectArticleList{}, err
	}

	field, err := cv.ValidateFieldsLikeOrCollect(in.Fileds)
	if err != nil {
		return OutputUserLikeCollectArticleList{}, err
	}

	return OutputUserLikeCollectArticleList{
		UserID:   userID,
		Page:     page,
		PageSize: pagesize,
		Fileds:   field,
	}, nil
}

func (cv *CommentValidate) ValidateFieldsLikeOrCollect(field string) (string, error) {
	fd := strings.TrimSpace(field)
	if fd == "" {
		return "", ErrFieldEmpty
	}
	if fd != "like" && fd != "collect" {
		return "", ErrFieldInvalid
	}
	return fd, nil
}

func (cv *CommentValidate) ValidateCommentParentID(parentID int64) (int64, error) {
	var pid = parentID
	if pid <= 0 {
		return 0, nil
	}
	pid, err := cv.ValidateSnowID(pid)
	if err != nil {
		return 0, ErrParentIDInvalid
	}
	return pid, nil
}

func (cv *CommentValidate) ValidateCommentAddress(address int64) (int64, error) {
	as := address
	if as <= 0 {
		return 0, ErrAddressEmpty
	}
	if as == 3489673482 {
		return as, nil
	}
	// 其他情况为评价文章 内容为文章ID 校验文章ID
	as, err := cv.ValidateSnowID(as)
	if err != nil {
		return 0, ErrAddressInvalid
	}
	return as, nil
}

func (cv *CommentValidate) ValidateCommentContent(content string) (string, error) {
	ct := strings.TrimSpace(content)
	if ct == "" {
		return "", ErrContentEmpty
	}
	return ct, nil
}
