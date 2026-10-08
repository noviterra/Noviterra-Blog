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

type GetArticleDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetArticleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleDetailLogic {
	return &GetArticleDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetArticleDetailLogic) GetArticleDetail(req *types.GetArticleDetailRequest) (resp *types.GetArticleDetailResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.GetArticleDetailResponse)

	if err := Validate.Struct(*req); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) && len(errs) > 0 {
			resp.Resp.Status = http.StatusBadRequest
			resp.Resp.Code = 1
			resp.Resp.Message = TypeName[errs[0].Field()] + "格式错误！"
		}
		return resp, nil
	}

	rep, err := l.svcCtx.ArticleModel.GetArticleDetail(l.ctx, &articlerpc.GetArticleDetailRequest{
		ArticleID: req.ArticleID,
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

	// 文章内容映射
	resp.Article.Title = rep.A.Title
	resp.Article.Cover = rep.A.Cover
	resp.Article.Content = rep.A.Content
	resp.Article.Information = rep.A.Information
	resp.Article.Tags = rep.A.Tags
	resp.Article.Status = 1
	resp.Article.CreatedAt = l.svcCtx.CommonUtils.ToStringTime(rep.A.CreatedAt)
	resp.Article.UpdatedAt = l.svcCtx.CommonUtils.ToStringTime(rep.A.UpdatedAt)
	resp.Article.ArticleID = rep.A.ID
	resp.Article.PublisherID = rep.A.PublisherID

	resp.Resp.Code = 0
	resp.Resp.Message = rep.Resp.Message
	resp.Resp.Status = http.StatusOK

	return resp, nil
}
