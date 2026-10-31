package middleware

import (
	"fmt"
	"net/http"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type BrowseMiddleware struct {
}

func (bm *BrowseMiddleware) BrowseMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var articlePath struct {
				Article string `path:"article"`
			}
			if err := httpx.ParseForm(r, &articlePath); err != nil {
				httpx.ErrorCtx(r.Context(), w, fmt.Errorf("路径异常:%w", err))
				return
			}

			// 开始记录文章的浏览次数
		}
	}
}
