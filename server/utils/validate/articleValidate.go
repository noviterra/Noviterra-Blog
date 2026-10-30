package validate

import (
	"Noviterra/server/utils/common"
	"errors"
	"strings"
	"time"
)

type ArticleValidate struct {
	commonUtils *common.CommonUtils
}

func NewArticleValidate(commonUtils *common.CommonUtils) *ArticleValidate {
	return &ArticleValidate{
		commonUtils: commonUtils,
	}
}

var (
	ErrArticleIDEmpty = errors.New("文章ID输入为空！")
	ErrArticleInvalid = errors.New("文章ID格式错误！")
)

const (
	epoch          = int64(1735689600000)
	timestampShift = 22
)

func (av *ArticleValidate) ValidateArticleID(articleID string) (string, error) {
	aid := strings.TrimSpace(articleID)
	if aid == "" {
		return "", ErrArticleIDEmpty
	}

	// ArticleID实际是雪花ID 先将ArticleID转换为int64
	sid, err := av.commonUtils.ToInt64(articleID)
	if err != nil {
		return "", ErrArticleInvalid
	}

	timeStamp := sid >> timestampShift
	now := time.Now().UnixMilli() - epoch
	if timeStamp <= 0 || timeStamp > now+5*60*1000 {
		return "", ErrArticleInvalid
	}

	return aid, nil
}
