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

func UpdateArticleCoverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateArticleCoverRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 获取上传的文件
		file, header, err := r.FormFile("cover")
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, &types.UpdateArticleCoverResponse{
				Cover: "",
				Resp: types.Response{
					Status:  http.StatusBadRequest,
					Code:    0,
					Message: "请上传图片",
				},
			})
			return
		}
		defer file.Close()

		l := logic.NewUpdateArticleCoverLogic(r.Context(), svcCtx)
		resp, err := l.UpdateArticleCover(&req, header, file)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
