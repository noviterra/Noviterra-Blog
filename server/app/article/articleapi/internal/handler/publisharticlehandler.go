// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"Noviterra/server/app/article/articleapi/internal/logic"
	"Noviterra/server/app/article/articleapi/internal/svc"
	"Noviterra/server/app/article/articleapi/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func PublishArticleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublishArticleRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewPublishArticleLogic(r.Context(), svcCtx)
		resp, err := l.PublishArticle(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
