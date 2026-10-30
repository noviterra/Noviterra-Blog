// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"Noviterra/server/app/identity/identityapi/internal/logic"
	"Noviterra/server/app/identity/identityapi/internal/svc"
	"Noviterra/server/app/identity/identityapi/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func loginByEmailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginByEmailRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewLoginByEmailLogic(r.Context(), svcCtx)
		resp, err := l.LoginByEmail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
