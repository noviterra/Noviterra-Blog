// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
	"Noviterra/server/utils/errcode"
	"context"
	"errors"
	"net/http"

	"Noviterra/server/app/article/articleapi/internal/svc"
	"Noviterra/server/app/article/articleapi/internal/types"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type getArticleDetailArticleModel interface {
	GetArticleDetail(ctx context.Context, in *articlerpc.GetArticleDetailRequest, opts ...grpc.CallOption) (*articlerpc.GetArticleDetailResponse, error)
}

type getArticleDetailServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
	GetSkipTokenInConf() string
}

type getArticleDetailCommonUtils interface {
	ToStringTime(ts int64) string
}

type GetArticleDetailLogic struct {
	ctx                     context.Context
	articleModel            getArticleDetailArticleModel
	serviceTokenInterceptor getArchiveServiceTokenInterceptor
	commonUtils             getArticleDetailCommonUtils
}

func NewGetArticleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleDetailLogic {
	return &GetArticleDetailLogic{
		ctx:                     ctx,
		articleModel:            svcCtx.ArticleModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
		commonUtils:             svcCtx.CommonUtils,
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
			return l.failGetArticleDetailResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return l.failGetArticleDetailResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	// 对于api层不需要鉴权的接口 也要将rpc服务凭证传入到metadata中
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"skip-token", l.serviceTokenInterceptor.GetSkipTokenInConf(),
	)

	rep, err := l.articleModel.GetArticleDetail(ctx, &articlerpc.GetArticleDetailRequest{
		ArticleID: req.ArticleID,
	})
	if err != nil {
		return l.failGetArticleDetailResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return l.failGetArticleDetailResponse(resp, resp.Resp.Message, errcode.InvalidParam, int(resp.Resp.Code))
	}

	return l.successGetArticleDetailResponse(resp, rep)
}

func (l *GetArticleDetailLogic) failGetArticleDetailResponse(resp *types.GetArticleDetailResponse, msg string, status, code int) (*types.GetArticleDetailResponse, error) {
	resp.Article = types.Article{}

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: msg,
	}

	return resp, nil
}

func (l *GetArticleDetailLogic) failGetArticleDetailResponseWithErr(resp *types.GetArticleDetailResponse, err error, status, code int) (*types.GetArticleDetailResponse, error) {
	resp.Article = types.Article{}

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "服务错误：" + err.Error(),
	}

	return resp, nil
}

func (l *GetArticleDetailLogic) successGetArticleDetailResponse(resp *types.GetArticleDetailResponse, rep *articlerpc.GetArticleDetailResponse) (*types.GetArticleDetailResponse, error) {
	// 数据映射
	resp.Article.Title = rep.A.Title
	resp.Article.Cover = rep.A.Cover
	resp.Article.Content = rep.A.Content
	resp.Article.Information = rep.A.Information
	resp.Article.Tags = rep.A.Tags
	resp.Article.Status = 1
	resp.Article.CreatedAt = l.commonUtils.ToStringTime(rep.A.CreatedAt)
	resp.Article.UpdatedAt = l.commonUtils.ToStringTime(rep.A.UpdatedAt)
	resp.Article.ArticleID = rep.A.ArticleID

	resp.Resp = types.Response{
		Status:  errcode.RequestSuccess,
		Code:    errcode.Success,
		Message: "文章加载完成！",
	}

	return resp, nil
}
