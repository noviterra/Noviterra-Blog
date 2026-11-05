package logic

import (
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"go.uber.org/zap"
)

type (
	deleteArticleCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	deleteArticleLogger interface {
		Error(msg string, fields ...zap.Field)
		Info(msg string, fields ...zap.Field)
		Warn(msg string, fields ...zap.Field)
	}

	deleteArticleResponse interface {
		SuccessResponse(resp *articlerpc.Response, msg string) (*articlerpc.Response, error)
		FailResponse(resp *articlerpc.Response, msg string, status int32, code int64) (*articlerpc.Response, error)
		FailResponseWithErr(resp *articlerpc.Response, err error) (*articlerpc.Response, error)
	}

	deleteArticleValidate interface {
		ValidateSnowID(id int64) (int64, error)
	}
	deleteArticleIdentityRepository interface {
		GetUserRole(id string) (int, error)
	}

	deleteArticleRepository interface {
		GetPublisherIDByArticleID(ctx context.Context, articleID int64) (string, error)
		DeleteArticle(ctx context.Context, articleID int64, deleterID string) (bool, error)
	}
)

type DeleteArticleLogic struct {
	ctx                context.Context
	ctxdata            deleteArticleCtxData
	logger             deleteArticleLogger
	articleResponse    deleteArticleResponse
	articleValidate    deleteArticleValidate
	articleRepository  deleteArticleRepository
	identityRepository deleteArticleIdentityRepository
}

func NewDeleteArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteArticleLogic {
	return &DeleteArticleLogic{
		ctx:                ctx,
		ctxdata:            svcCtx.CtxData,
		logger:             svcCtx.Logger,
		articleResponse:    svcCtx.ArticleResponse,
		articleValidate:    svcCtx.ArticleValidate,
		articleRepository:  svcCtx.ArticleRepository,
		identityRepository: svcCtx.IdentityRepository,
	}
}

// 删除文章
func (l *DeleteArticleLogic) DeleteArticle(in *articlerpc.DeleteArticleRequest) (*articlerpc.Response, error) {
	// todo: add your logic here and delete this line
	resp := new(articlerpc.Response)
	// 删除文章 需要身份校验
	deleterID := l.ctxdata.GetUserIDFromContext(l.ctx)
	if deleterID == "" {
		l.logger.Error("article/deletearticlelogic：用户ID为空，疑似绕过api层进行请求（暂未锁定IP）")
		return l.articleResponse.FailResponse(resp, "用户ID为空，请正常访问资源！", errcode.Unauthorized, errcode.Failed)
	}

	// 校验发起删除请求者的用户权限 至少管理员能操作
	deleterRole, err := l.identityRepository.GetUserRole(deleterID)
	if deleterRole < 1 { // 1管理员 2超级管理员
		l.logger.Error("article/deletearticle：普通用户访问删除文章操作，需要严重标记该用户", zap.String("deleterID", deleterID), zap.Int("role", deleterRole))
		return l.articleResponse.FailResponse(resp, "权限不足", errcode.Forbidden, errcode.Failed)
	}

	// 校验文章ID
	articleID, err := l.articleValidate.ValidateSnowID(in.ArticleID)
	if err != nil {
		l.logger.Error("article/deletearticle：校验文章ID失败", zap.Int64("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.FailResponseWithErr(resp, err)
	}
	if articleID == 0 {
		l.logger.Error("article/deletearticle：校验文章ID错误，不存在的文章ID", zap.Int64("ArticleID", in.ArticleID))
		return l.articleResponse.FailResponse(resp, "文章不存在！", errcode.NotFound, errcode.Failed)
	}

	// 根据文章ID获取文章发布者（此操作也可以侧面测出文章ID是否存在） 并得到文章发布者的身份权限
	publisherID, err := l.articleRepository.GetPublisherIDByArticleID(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("article/deletearticle：数据库根据文章ID查询发布者ID操作失败", zap.Int64("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.FailResponse(resp, "数据库操作失败！", errcode.ServerError, errcode.Failed)
	}
	if publisherID == "" {
		l.logger.Error("article/deletearticle：数据库根据文章ID查询到的发布者ID为空", zap.Int64("Article", in.ArticleID))
		return l.articleResponse.FailResponse(resp, "要删除的文章不存在！", errcode.NotFound, errcode.Failed)
	}

	// 判断删除者与发布者是否是同一个
	if deleterID != publisherID { // 不是同一个则对比权限
		// 获取发布者的身份权限
		publisherRole, err := l.identityRepository.GetUserRole(publisherID)
		if err != nil {
			l.logger.Error("article/deletearticle：数据库查询发布者身份权限失败", zap.String("PublisherID", publisherID), zap.String("ERROR", err.Error()))
			return l.articleResponse.FailResponse(resp, "数据库操作失败！", errcode.ServerError, errcode.Failed)
		}
		if deleterRole <= publisherRole { // 相同权限不可删除 只有高一级才可执行
			l.logger.Error("article/deletearticle：删除操作者的身份权限不高于发布者权限，删除操作失败，请求已拦截",
				zap.String("PublisherID", publisherID), zap.Int("PublisherRole", publisherRole),
				zap.String("DeleterID", deleterID), zap.Int("DeleterRole", deleterRole))
			return l.articleResponse.FailResponse(resp, "权限不高于发布者，删除失败！", errcode.Forbidden, errcode.Failed)
		}
		// 过滤得到的是 删除者的权限大于发布者
	}
	// 执行删除操作
	deleteOK, err := l.articleRepository.DeleteArticle(l.ctx, in.ArticleID, deleterID)
	if err != nil {
		l.logger.Error("article/deletearticle：数据库删除文章失败", zap.Int64("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.FailResponse(resp, "数据库操作失败！", errcode.ServerError, errcode.Failed)
	}
	if !deleteOK {
		l.logger.Error("article/deletearticle：数据库删除文章数量为0，请查库检查", zap.Int64("ArticleID", in.ArticleID))
	}

	l.logger.Info("article/deletearticle：文章删除成功", zap.Int64("ArticleID", in.ArticleID), zap.String("DeleterID", deleterID))
	return l.articleResponse.SuccessResponse(resp, "文章删除成功！")
}
