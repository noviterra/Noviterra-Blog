// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"

	"Noviterra/server/app/article/articleapi/internal/svc"
	"Noviterra/server/app/article/articleapi/internal/types"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"github.com/go-playground/validator/v10"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type publishArticleCommonUtils interface {
	SaveImageFile(header *multipart.FileHeader, file multipart.File, fileDie string, fileName string) (string, error)
}
type PublishArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext

	commonUtils publishArticleCommonUtils
}

func NewPublishArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishArticleLogic {
	return &PublishArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,

		commonUtils: svcCtx.CommonUtils,
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
			resp.Status = http.StatusBadRequest
			resp.Code = 1
			resp.Message = TypeName[errs[0].Field()] + "格式错误！"
		}
		return resp, nil
	}

	// 从上下文中获取用户ID
	publisherID := l.svcCtx.Middleware.GetUserIDFromContext(l.ctx)
	if publisherID == "" {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "用户ID为空！"
		return resp, nil
	}

	// 获取用户身份权限
	role := l.svcCtx.Middleware.GetUserRoleFromContext(l.ctx)
	if role != 1 { // 1为管理员
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "用户身份权限不足！"
		return resp, nil
	}

	// 将用户信息注入metadata
	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.svcCtx.Config.Interceptor.ServiceToken,
		"user-id", publisherID,
	)

	// 调用rpc服务方法 执行核心业务
	rep, err := l.svcCtx.ArticleModel.PublishArticle(ctx, &articlerpc.PublishArticleRequest{
		Tags:        req.Tags,
		Title:       req.Title,
		Content:     req.Content,
		Information: req.Information,
	})
	if err != nil {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "ERROR:" + err.Error()
		return resp, nil
	}
	if rep.Resp.Code != 0 {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = rep.Resp.Message
		return resp, nil
	}

	// 文件校验 图片保存的名称格式为  cover*time*.* 保证进入数据库更新操作时rowAffected不为0
	cover, err := l.commonUtils.SaveImageFile(header, file, "articlecover", rep.ArticleID)
	if err != nil {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "ERROR:" + err.Error()
		return resp, nil
	}
	if cover == "" {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "封面为空！"
		return resp, nil
	}
	// 更新cover字段  使用注入了metadata数据的ctx
	rep2, err := l.svcCtx.ArticleModel.UpdateArticleCover(ctx, &articlerpc.UpdateArticleCoverRequest{
		NewCover:  cover,
		ArticleID: rep.ArticleID,
	})
	if err != nil {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = "ERROR:" + err.Error()
		return resp, nil
	}
	if rep2.Resp.Code != 0 {
		resp.Status = http.StatusBadRequest
		resp.Code = 1
		resp.Message = rep2.Resp.Message
		return resp, nil
	}

	resp.Code = 0
	resp.Message = "文章发布成功！"
	resp.Status = http.StatusOK

	return resp, nil
}
