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

type publishArticleModel interface {
	PublishArticle(ctx context.Context, in *articlerpc.PublishArticleRequest, opts ...grpc.CallOption) (*articlerpc.PublishArticleResponse, error)
	UpdateArticleCover(ctx context.Context, in *articlerpc.UpdateArticleCoverRequest, opts ...grpc.CallOption) (*articlerpc.UpdateArticleCoverResponse, error)
}

type publishArticleValidate interface {
	SaveImageFile(header *multipart.FileHeader, file multipart.File, fileDie string, articleID int64) (string, error)
}

type publishArticlectxdata interface {
	GetUserIDFromContext(ctx context.Context) string
	GetUserRoleFromContext(ctx context.Context) int
}

type publishArticleServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
}
type PublishArticleLogic struct {
	ctx                     context.Context
	articleModel            publishArticleModel
	commentValidate         publishArticleValidate
	ctxdata                 publishArticlectxdata
	serciceTokenInterceptor publishArticleServiceTokenInterceptor
}

func NewPublishArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishArticleLogic {
	return &PublishArticleLogic{
		ctx:                     ctx,
		articleModel:            svcCtx.ArticleModel,
		commentValidate:         svcCtx.ArticleValidate,
		ctxdata:                 svcCtx.CtxData,
		serciceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *PublishArticleLogic) PublishArticle(req *types.PublishArticleRequest,
	header *multipart.FileHeader, file multipart.File) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	resp = new(types.Response)

	// 参数格式校验
	if err := Validate.Struct(*req); err != nil {
		var errs validator.ValidationErrors
		if errors.As(err, &errs) && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	// 从上下文中获取用户ID
	publisherID := l.ctxdata.GetUserIDFromContext(l.ctx)
	if publisherID == "" {
		return failResponse(resp, "用户ID为空！", errcode.InvalidParam, errcode.Failed)
	}

	// 获取用户身份权限
	role := l.ctxdata.GetUserRoleFromContext(l.ctx)
	if role != 1 { // 1为管理员
		return failResponse(resp, "用户身份权限不足！", errcode.InvalidParam, errcode.Failed)
	}

	// 将用户信息注入metadata
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serciceTokenInterceptor.GetServiceTokenInConf(),
		"user-id", publisherID,
	)

	// 调用rpc服务方法 执行发布文章核心业务
	rep, err := l.articleModel.PublishArticle(ctx, &articlerpc.PublishArticleRequest{
		Tags:        req.Tags,
		Title:       req.Title,
		Content:     req.Content,
		Information: req.Information,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Resp.Code != 0 {
		return failResponse(resp, rep.Resp.Message, errcode.InvalidParam, int(rep.Resp.Code))
	}

	// 文件校验 图片保存的名称格式为  cover*time*.* 保证进入数据库更新操作时rowAffected不为0
	cover, err := l.commentValidate.SaveImageFile(header, file, "articlecover", rep.ArticleID)
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if cover == "" {
		return failResponse(resp, "封面为空！", errcode.InvalidParam, errcode.Failed)
	}
	// 更新cover字段  使用注入了metadata数据的ctx
	rep2, err := l.articleModel.UpdateArticleCover(ctx, &articlerpc.UpdateArticleCoverRequest{
		NewCover:  cover,
		ArticleID: rep.ArticleID,
	})
	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep2.Resp.Code != 0 {
		return failResponse(resp, "文章上传成功但封面上传失败！", errcode.InvalidParam, errcode.Failed)
	}

	return successResponse(resp, "文章发布成功！")
}
