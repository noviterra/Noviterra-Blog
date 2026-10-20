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
		// 先解析multipart 普通字段进入r.Form，后续httpx.Parse才能读到
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			httpx.OkJsonCtx(r.Context(), w, &types.Response{
				Status:  http.StatusBadRequest,
				Code:    1,
				Message: "表单解析失败：" + err.Error(),
			})
			return
		}

		var req types.PublishArticleRequest
		if err := httpx.ParseForm(r, &req); err != nil {
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

		l := logic.NewPublishArticleLogic(r.Context(), svcCtx)
		resp, err := l.PublishArticle(&req, header, file)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
