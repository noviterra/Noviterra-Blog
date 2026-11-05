// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/article/articleapi/internal/svc"
	"Noviterra/server/app/article/articleapi/internal/types"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type (
	getUserLikeCollectArticleModel interface {
		GetUserLikeArticleList(ctx context.Context,
			in *articlerpc.GetUserLikeCollectArticleListRequest, opts ...grpc.CallOption) (*articlerpc.GetArticleListResponse, error)
	}

	getUserLikeCollectArticleServiceTokenInterceptor interface {
		GetServiceTokenInConf() string
		GetSkipTokenInConf() string
	}
)

type GetUserLikeCollectArticleListLogic struct {
	ctx                     context.Context
	articleModel            getUserLikeCollectArticleModel
	serviceTokenInterceptor getUserLikeCollectArticleServiceTokenInterceptor
}

// 查看用户点赞或收藏文章列表
func NewGetUserLikeCollectArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLikeCollectArticleListLogic {
	return &GetUserLikeCollectArticleListLogic{
		ctx:                     ctx,
		articleModel:            svcCtx.ArticleModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *GetUserLikeCollectArticleListLogic) GetUserLikeCollectArticleList(req *types.GetUserLikeCollectListRequest) (resp *types.GetArticleListResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.GetArticleListResponse)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failGetArticleListResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failGetArticleListResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"skip-token", l.serviceTokenInterceptor.GetSkipTokenInConf(),
	)

	rep, err := l.articleModel.GetUserLikeArticleList(ctx, &articlerpc.GetUserLikeCollectArticleListRequest{
		UserID:   req.UserID,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		Field:    req.Fields,
	})

	if err != nil {
		return failGetArticleListResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return failGetArticleListResponse(resp, rep.Resp.Message, int(rep.Resp.Status), int(rep.Resp.Code))
	}

	return successGetArticleListResponse(resp, rep)
}

func failGetArticleListResponse(resp *types.GetArticleListResponse,
	msg string, status, code int) (*types.GetArticleListResponse, error) {
	if resp == nil {
		resp = new(types.GetArticleListResponse)
	}
	resp.ArticlesList = nil
	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: msg,
	}

	return resp, nil
}

func failGetArticleListResponseWithErr(resp *types.GetArticleListResponse,
	err error, status, code int) (*types.GetArticleListResponse, error) {
	if resp == nil {
		resp = new(types.GetArticleListResponse)
	}
	resp.ArticlesList = nil
	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "请求失败：" + err.Error(),
	}

	return resp, nil
}

func successGetArticleListResponse(resp *types.GetArticleListResponse,
	rep *articlerpc.GetArticleListResponse) (*types.GetArticleListResponse, error) {
	if resp == nil {
		resp = new(types.GetArticleListResponse)
	}
	// 数据映射
	if len(rep.Articles) > 0 {
		for _, article := range rep.Articles {
			var a types.ArticleList
			a.ArticleID = article.ArticleID
			a.Title = article.Title
			a.Cover = article.Cover
			a.Information = article.Information
			a.Tags = article.Tags
			a.CreatedAt = article.CreatedAt

			resp.ArticlesList = append(resp.ArticlesList, a)
		}
	} else {
		resp.ArticlesList = []types.ArticleList{}
	}

	resp.Resp = types.Response{
		Status:  int(rep.Resp.Status),
		Code:    int(rep.Resp.Code),
		Message: rep.Resp.Message,
	}

	return resp, nil
}
