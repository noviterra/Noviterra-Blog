// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"Noviterra/server/app/notice/noticeapi/internal/logic"
	"Noviterra/server/app/notice/noticeapi/internal/svc"
	"Noviterra/server/app/notice/noticeapi/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 发布公告
func PublishNoticeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PublishStationNoticeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewPublishNoticeLogic(r.Context(), svcCtx)
		resp, err := l.PublishNotice(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
