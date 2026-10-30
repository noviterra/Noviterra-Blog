// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"Noviterra/server/app/notice/noticeapi/internal/logic"
	"Noviterra/server/app/notice/noticeapi/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 得到公告
func GetNoticeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetNoticeLogic(r.Context(), svcCtx)
		resp, err := l.GetNotice()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
