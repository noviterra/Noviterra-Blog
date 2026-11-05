package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type (
	browseMiddlewareCommentCache interface {
		View(ctx context.Context, articleID int64) error
	}

	browseMiddlewareArticleRepository interface {
		CheckArticleExist(ctx context.Context, articleID int64) (bool, error)
	}

	browseMiddlewareArticleValidate interface {
		ValidateSnowID(articleID int64) (int64, error)
	}
)

type BrowseMiddleware struct {
	commentCache      browseMiddlewareCommentCache
	articleRepository browseMiddlewareArticleRepository
	articleValidate   browseMiddlewareArticleValidate
}

func NewBrowseMiddleware(commentCache browseMiddlewareCommentCache,
	articleRepository browseMiddlewareArticleRepository,
	articleValidate browseMiddlewareArticleValidate) *BrowseMiddleware {
	return &BrowseMiddleware{
		commentCache:      commentCache,
		articleRepository: articleRepository,
		articleValidate:   articleValidate,
	}
}

func (bm *BrowseMiddleware) BrowseMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var articlePath struct {
				ArticleID int64 `path:"article"`
			}
			if err := httpx.ParseForm(r, &articlePath); err != nil {
				httpx.ErrorCtx(r.Context(), w, fmt.Errorf("路径异常:%w", err))
				return
			}

			// 判断路径参数中的ID是否存在
			// 格式校验
			articleID, err := bm.articleValidate.ValidateSnowID(articlePath.ArticleID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, fmt.Errorf("路径参数有误：%w", err))
				return
			}
			// 查库
			exist, err := bm.articleRepository.CheckArticleExist(context.Background(), articleID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, fmt.Errorf("数据库操作异常；%w", err))
				return
			}
			if !exist {
				httpx.ErrorCtx(r.Context(), w, fmt.Errorf("文章不存在！"))
				return
			}

			// 开始记录文章的浏览次数
			err = bm.commentCache.View(context.Background(), articleID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, fmt.Errorf("缓存操作异常；%w", err))
				return
			}

			next(w, r)
		}
	}
}
