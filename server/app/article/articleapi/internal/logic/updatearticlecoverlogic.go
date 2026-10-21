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
	"mime/multipart"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type updateArticleCoverArticleModel interface {
	UpdateArticleCover(ctx context.Context, in *articlerpc.UpdateArticleCoverRequest, opts ...grpc.CallOption) (*articlerpc.UpdateArticleCoverResponse, error)
}

type updateArticleCoverAuthMiddleware interface {
	GetUserIDFromContext(ctx context.Context) string
	GetUserRoleFromContext(ctx context.Context) int
}

type updateArticleCoverServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
}

type updatearticlecoverCommonutils interface {
	SaveImageFile(header *multipart.FileHeader, file multipart.File, saveDir, fileName string) (string, error)
}

type UpdateArticleCoverLogic struct {
	ctx                     context.Context
	articleModel            updateArticleCoverArticleModel
	commonUtils             updatearticlecoverCommonutils
	authMiddleware          updateArticleCoverAuthMiddleware
	serviceTokenInterceptor updateArticleCoverServiceTokenInterceptor
}

func NewUpdateArticleCoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateArticleCoverLogic {
	return &UpdateArticleCoverLogic{
		ctx:                     ctx,
		articleModel:            svcCtx.ArticleModel,
		commonUtils:             svcCtx.CommonUtils,
		authMiddleware:          svcCtx.AuthMiddleware,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *UpdateArticleCoverLogic) UpdateArticleCover(req *types.UpdateArticleCoverRequest,
	header *multipart.FileHeader, file multipart.File) (resp *types.UpdateArticleCoverResponse, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.UpdateArticleCoverResponse)

	if err := Validate.Struct(*req); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) && len(errs) > 0 {
			return l.failUpdateArticleCoverResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return l.failUpdateArticleCoverResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	// 从上下文中获取用户ID
	publisherID := l.authMiddleware.GetUserIDFromContext(l.ctx)
	if publisherID == "" {
		return l.failUpdateArticleCoverResponse(resp, "用户ID为空！", errcode.InvalidParam, errcode.Failed)
	}

	// 获取用户身份权限
	role := l.authMiddleware.GetUserRoleFromContext(l.ctx)
	if role < 1 {
		return l.failUpdateArticleCoverResponse(resp, "用户ID为空！", errcode.InvalidParam, errcode.Failed)
	}

	// 信息注入到metadata中
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", publisherID,
	)

	// 文件校验 图片保存的名称格式为  cover*time*.*
	cover, err := l.commonUtils.SaveImageFile(header, file, "articlecover", req.ArticleID)
	if err != nil {
		return l.failUpdateArticleCoverResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if cover == "" {
		return l.failUpdateArticleCoverResponse(resp, "封面为空！", errcode.InvalidParam, errcode.Failed)
	}

	// 调用rpc服务方法 执行核心业务
	rep, err := l.articleModel.UpdateArticleCover(ctx, &articlerpc.UpdateArticleCoverRequest{
		ArticleID: req.ArticleID,
		NewCover:  cover,
	})
	if err != nil {
		return l.failUpdateArticleCoverResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return l.failUpdateArticleCoverResponse(resp, rep.Resp.Message, errcode.InvalidParam, int(rep.Resp.Code))
	}

	return l.successUpdateArticleCoverResponse(resp, rep)
}

func (l *UpdateArticleCoverLogic) failUpdateArticleCoverResponse(resp *types.UpdateArticleCoverResponse, msg string, status, code int) (*types.UpdateArticleCoverResponse, error) {
	resp.Cover = ""

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: msg,
	}

	return resp, nil
}

func (l *UpdateArticleCoverLogic) failUpdateArticleCoverResponseWithErr(resp *types.UpdateArticleCoverResponse, err error, status, code int) (*types.UpdateArticleCoverResponse, error) {
	resp.Cover = ""

	resp.Resp = types.Response{
		Status:  status,
		Code:    code,
		Message: "服务失败：" + err.Error(),
	}

	return resp, nil
}

func (l *UpdateArticleCoverLogic) successUpdateArticleCoverResponse(resp *types.UpdateArticleCoverResponse, rep *articlerpc.UpdateArticleCoverResponse) (*types.UpdateArticleCoverResponse, error) {
	resp.Cover = rep.Cover

	resp.Resp = types.Response{
		Status:  errcode.RequestSuccess,
		Code:    errcode.Success,
		Message: "封面更新完成！",
	}

	return resp, nil
}
