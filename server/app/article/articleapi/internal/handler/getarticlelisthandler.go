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

func GetArticleListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetArticleListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetArticleListLogic(r.Context(), svcCtx)
		resp, err := l.GetArticleList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
