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

type searchArticleModel interface {
	SearchArticle(ctx context.Context, in *articlerpc.SearchArticleRequest, opts ...grpc.CallOption) (*articlerpc.SearchArticleResponse, error)
}

type searchArticleServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
	GetSkipTokenInConf() string
}
type SearchArticleLogic struct {
	ctx                     context.Context
	articleModel            searchArticleModel
	serviceTokenInterceptor searchArticleServiceTokenInterceptor
}

func NewSearchArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchArticleLogic {
	return &SearchArticleLogic{
		ctx:                     ctx,
		articleModel:            svcCtx.ArticleModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *SearchArticleLogic) SearchArticle(req *types.SearchArticleRequest) (resp *types.SearchArticleResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.SearchArticleResponse)

	if err := Validate.Struct(*req); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) && len(errs) > 0 {
			return l.failSearchArticleResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return l.failSearchArticleResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	// 对于api层不需要鉴权的接口 也要将rpc服务凭证传入到metadata中
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"skip-token", l.serviceTokenInterceptor.GetSkipTokenInConf(),
	)

	rep, err := l.articleModel.SearchArticle(ctx, &articlerpc.SearchArticleRequest{
		Query:    req.Query,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return l.failSearchArticleResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return l.failSearchArticleResponse(resp, rep.Resp.Message, errcode.InvalidParam, int(rep.Resp.Code))
	}

	return l.successSearchArticleResponse(resp, rep)
}

func (l *SearchArticleLogic) failSearchArticleResponse(resp *types.SearchArticleResponse, msg string, status, code int) (*types.SearchArticleResponse, error) {
	resp.ArticlesList = nil

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: msg,
	}

	return resp, nil
}

func (l *SearchArticleLogic) failSearchArticleResponseWithErr(resp *types.SearchArticleResponse, err error, status, code int) (*types.SearchArticleResponse, error) {
	resp.ArticlesList = nil

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "服务错误：" + err.Error(),
	}

	return resp, nil
}

func (l *SearchArticleLogic) successSearchArticleResponse(resp *types.SearchArticleResponse, rep *articlerpc.SearchArticleResponse) (*types.SearchArticleResponse, error) {
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
		Message: "文章修改成功！",
	}

	return resp, nil
}
