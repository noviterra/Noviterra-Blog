// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"Noviterra/server/app/user/userapi/internal/logic"
	"Noviterra/server/app/user/userapi/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func getUserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetUserInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetUserInfo()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
