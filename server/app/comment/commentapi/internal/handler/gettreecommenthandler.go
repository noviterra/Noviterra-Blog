// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"Noviterra/server/app/comment/commentapi/internal/logic"
	"Noviterra/server/app/comment/commentapi/internal/svc"
	"Noviterra/server/app/comment/commentapi/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 展开某个评论
func GetTreeCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetTreeCommentRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetTreeCommentLogic(r.Context(), svcCtx)
		resp, err := l.GetTreeComment(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
