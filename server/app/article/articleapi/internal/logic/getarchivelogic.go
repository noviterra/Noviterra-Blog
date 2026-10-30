// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/article/articlerpc/types/articlerpc"
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/article/articleapi/internal/svc"
	"Noviterra/server/app/article/articleapi/internal/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type getArchiveServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
	GetSkipTokenInConf() string
}

type getArchiveArticleModel interface {
	GetArchive(ctx context.Context, in *articlerpc.GetArchiveRequest, opts ...grpc.CallOption) (*articlerpc.GetArchiveResponse, error)
}
type GetArchiveLogic struct {
	ctx                     context.Context
	articleModel            getArchiveArticleModel
	serviceTokenInterceptor getArchiveServiceTokenInterceptor
}

// 归档 时间轴展示博客与
func NewGetArchiveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArchiveLogic {
	return &GetArchiveLogic{
		ctx:                     ctx,
		articleModel:            svcCtx.ArticleModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *GetArchiveLogic) GetArchive() (resp *types.GetArchiveResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.GetArchiveResponse)

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"srevice-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"skip-token", l.serviceTokenInterceptor.GetSkipTokenInConf(),
	)

	// 调用核心业务函数
	rep, err := l.articleModel.GetArchive(ctx, &articlerpc.GetArchiveRequest{})
	if err != nil {
		resp.ArticleList = nil
		return l.failGetArchiveResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return l.failGetArchiveResponse(resp, rep.Resp.Message, errcode.InvalidParam, int(rep.Resp.Code))
	}

	return l.successGetArchiveResponse(resp, rep)
}

// 非基础响应体即非Response响应体 一律绑定到其下的结构体中

func (l *GetArchiveLogic) failGetArchiveResponse(resp *types.GetArchiveResponse, msg string, status, code int) (*types.GetArchiveResponse, error) {
	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: msg,
	}
	resp.ArticleList = nil

	return resp, nil
}

// 直接将错误拼接入msg 核心业务返回的err 都是自定义的
func (l *GetArchiveLogic) failGetArchiveResponseWithErr(resp *types.GetArchiveResponse, err error, status, code int) (*types.GetArchiveResponse, error) {
	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "服务错误：" + err.Error(),
	}
	resp.ArticleList = nil

	return resp, nil
}

func (l *GetArchiveLogic) successGetArchiveResponse(resp *types.GetArchiveResponse, rep *articlerpc.GetArchiveResponse) (*types.GetArchiveResponse, error) {
	resp.Resp = types.Response{
		Status:  errcode.RequestSuccess,
		Code:    errcode.Success,
		Message: "归档加载完成！",
	}

	// 数据映射
	result := make([]types.ArchiveArticle, 0)
	for _, archive := range rep.Archives {
		var a types.ArchiveArticle
		a.Title = archive.Title
		a.Action = archive.Action
		a.ActionTime = archive.Time

		result = append(result, a)
	}
	resp.ArticleList = result

	return resp, nil
}
