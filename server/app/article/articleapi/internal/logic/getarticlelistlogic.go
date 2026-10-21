// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/article/articleapi/internal/svc"
	"Noviterra/server/app/article/articleapi/internal/types"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
	"Noviterra/server/utils/errcode"
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type getArticleListArticleModel interface {
	GetArticleList(ctx context.Context, in *articlerpc.GetArticleListRequest, opts ...grpc.CallOption) (*articlerpc.GetArticleListResponse, error)
}

type getArticleListServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
	GetSkipTokenInConf() string
}

type GetArticleListLogic struct {
	ctx                     context.Context
	articleModel            getArticleListArticleModel
	serviceTokenInterceptor getArticleListServiceTokenInterceptor
}

func NewGetArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleListLogic {
	return &GetArticleListLogic{
		ctx:                     ctx,
		articleModel:            svcCtx.ArticleModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *GetArticleListLogic) GetArticleList(req *types.GetArticleListRequest) (resp *types.GetArticleListResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.GetArticleListResponse)

	if err := Validate.Struct(*req); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) && len(errs) > 0 {
			return l.failGetArticleListResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return l.failGetArticleListResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	// 对于api层不需要鉴权的接口 也要将rpc服务凭证传入到metadata中
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"skip-token", l.serviceTokenInterceptor.GetSkipTokenInConf(),
	)

	rep, err := l.articleModel.GetArticleList(ctx, &articlerpc.GetArticleListRequest{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return l.failGetArticleListResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return l.failGetArticleListResponse(resp, rep.Resp.Message, errcode.InvalidParam, int(rep.Resp.Code))
	}

	return l.successGetArticleListResponse(resp, rep)
}

func (l *GetArticleListLogic) failGetArticleListResponse(resp *types.GetArticleListResponse, msg string, status, code int) (*types.GetArticleListResponse, error) {
	resp.ArticlesList = nil

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: msg,
	}

	return resp, nil
}

func (l *GetArticleListLogic) failGetArticleListResponseWithErr(resp *types.GetArticleListResponse, err error, status, code int) (*types.GetArticleListResponse, error) {
	resp.ArticlesList = nil

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "服务失败：" + err.Error(),
	}

	return resp, nil
}

func (l *GetArticleListLogic) successGetArticleListResponse(resp *types.GetArticleListResponse, rep *articlerpc.GetArticleListResponse) (*types.GetArticleListResponse, error) {
	// 文章列表内容映射
	for _, a := range rep.Articles {
		var elem types.ArticleList
		elem.Title = a.Title
		elem.Cover = a.Cover
		elem.Tags = a.Tags
		elem.Information = a.Information

		resp.ArticlesList = append(resp.ArticlesList, elem)
	}

	resp.Resp = types.Response{
		Status:  errcode.RequestSuccess,
		Code:    errcode.Success,
		Message: "文章列表加载完成！",
	}

	return resp, nil
}
