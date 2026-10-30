// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"Noviterra/server/app/site/siteapi/internal/logic"
	"Noviterra/server/app/site/siteapi/internal/svc"
	"Noviterra/server/app/site/siteapi/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 友链申请
func ApplyFriendLinkHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FriendLinkRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewApplyFriendLinkLogic(r.Context(), svcCtx)
		resp, err := l.ApplyFriendLink(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
