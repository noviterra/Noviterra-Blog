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
	getUserArticleListArticleModel interface {
		GetUserArticleList(ctx context.Context,
			in *articlerpc.GetUserArticleListRequest, opts ...grpc.CallOption) (*articlerpc.GetArticleListResponse, error)
	}

	getUserArticleListServiceTokenInterceptor interface {
		GetServiceTokenInConf() string
		GetSkipTokenInConf() string
	}
)

type GetUsersArticleListLogic struct {
	ctx                     context.Context
	articleModel            getUserArticleListArticleModel
	serviceTokenInterceptor getUserArticleListServiceTokenInterceptor
}

// 查看用户发布文章列表
func NewGetUsersArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUsersArticleListLogic {
	return &GetUsersArticleListLogic{
		ctx:                     ctx,
		articleModel:            svcCtx.ArticleModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *GetUsersArticleListLogic) GetUsersArticleList(req *types.GetUsersArticleListRequest) (resp *types.GetArticleListResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.GetArticleListResponse)

	if err := Validate.Struct(*req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return l.failGetArticleListResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return l.failGetArticleListResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"skip-token", l.serviceTokenInterceptor.GetSkipTokenInConf(),
	)

	rep, err := l.articleModel.GetUserArticleList(ctx, &articlerpc.GetUserArticleListRequest{
		UserID:   req.UserID,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return l.failGetArticleListResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return l.failGetArticleListResponse(resp, rep.Resp.Message, int(rep.Resp.Status), int(rep.Resp.Code))
	}

	return l.successGetArticleListResponse(resp, rep)
}

func (l *GetUsersArticleListLogic) failGetArticleListResponse(resp *types.GetArticleListResponse,
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

func (l *GetUsersArticleListLogic) failGetArticleListResponseWithErr(resp *types.GetArticleListResponse,
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

func (l *GetUsersArticleListLogic) successGetArticleListResponse(resp *types.GetArticleListResponse,
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
