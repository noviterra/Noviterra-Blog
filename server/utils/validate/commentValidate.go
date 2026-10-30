package validate

import (
	"errors"
	"strings"
)

type CommentValidate struct {
	*ArticleValidate
	*IdentityValidate
}

func NewCommentValidate(articleValidate *ArticleValidate,
	identityValidate *IdentityValidate) *CommentValidate {
	return &CommentValidate{
		ArticleValidate:  articleValidate,
		IdentityValidate: identityValidate,
	}
}

var (
	ErrActionEmpty   = errors.New("点赞操作内容为空！")
	ErrActionInvalid = errors.New("点赞操作类型有误！")
)

type InputLikeComment struct {
	ArticleID string
	UserID    string
	Action    string
}

type OutputLikeComment struct {
	ArticleID string
	UserID    string
	Action    string
}

func (cv *CommentValidate) ValidateLikeComment(in InputLikeComment, field string) (OutputLikeComment, error) {
	articleID, err := cv.ValidateArticleID(in.ArticleID)
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
