// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"Noviterra/server/app/user/userapi/internal/logic"
	"Noviterra/server/app/user/userapi/internal/svc"
	"Noviterra/server/app/user/userapi/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func sendEmailVerificationCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendEmailVerificationCodeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewSendEmailVerificationCodeLogic(r.Context(), svcCtx)
		resp, err := l.SendEmailVerificationCode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
