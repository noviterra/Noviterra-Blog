// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
	"context"
	"errors"
	"net/http"

	"Noviterra/server/app/article/articleapi/internal/svc"
	"Noviterra/server/app/article/articleapi/internal/types"

	"github.com/go-playground/validator/v10"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetArticleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleListLogic {
	return &GetArticleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetArticleListLogic) GetArticleList(req *types.GetArticleListRequest) (resp *types.GetArticleListResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.GetArticleListResponse)

	if err := Validate.Struct(*req); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) && len(errs) > 0 {
			resp.Resp.Status = http.StatusBadRequest
			resp.Resp.Code = 1
			resp.Resp.Message = TypeName[errs[0].Field()] + "格式错误！"
		}
		return resp, nil
	}

	rep, err := l.svcCtx.ArticleModel.GetArticleList(l.ctx, &articlerpc.GetArticleListRequest{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = "ERROR:" + err.Error()
		return resp, nil
	}
	if rep.Resp.Code != 0 {
		resp.Resp.Status = http.StatusBadRequest
		resp.Resp.Code = 1
		resp.Resp.Message = rep.Resp.Message
		return resp, nil
	}

	// 文章列表内容映射
	for _, a := range rep.Articles {
		var elem types.ArticleList
		elem.Title = a.Title
		elem.Cover = a.Cover
		elem.Tags = a.Tags
		elem.Information = a.Information

		resp.ArticlesList = append(resp.ArticlesList, elem)
	}

	resp.Resp.Code = 0
	resp.Resp.Message = rep.Resp.Message
	resp.Resp.Status = http.StatusOK

	return resp, nil
}
