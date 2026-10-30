package logic

import (
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"go.uber.org/zap"
)

type deleteArticleCtxData interface {
	GetUserIDFromContext(ctx context.Context) string
}

type deleteArticleLogger interface {
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
}

type deleteArticleResponse interface {
	RPCSuccessResponse(msg string) (*articlerpc.Response, error)
	RPCFailResponse(msg string, code int64) (*articlerpc.Response, error)
	RPCFailResponseWithErr(err error) (*articlerpc.Response, error)
}

type deleteArticleCommonUtils interface {
	CheckArticleID(articleID string) (bool, error)
}

type deleteArticleIdentityRepository interface {
	GetUserRole(id string) (int, error)
}

type deleteArticleRepository interface {
	GetPublisherIDByArticleID(ctx context.Context, articleID string) (string, error)
	DeleteArticle(ctx context.Context, articleID string, deleterID string) (bool, error)
}

type DeleteArticleLogic struct {
	ctx                context.Context
	ctxdata            deleteArticleCtxData
	logger             deleteArticleLogger
	articleResponse    deleteArticleResponse
	commonUtils        deleteArticleCommonUtils
	articleRepository  deleteArticleRepository
	identityRepository deleteArticleIdentityRepository
}

func NewDeleteArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteArticleLogic {
	return &DeleteArticleLogic{
		ctx:                ctx,
		ctxdata:            svcCtx.CtxData,
		logger:             svcCtx.Logger,
		articleResponse:    svcCtx.ArticleResponse,
		commonUtils:        svcCtx.CommonUtils,
		articleRepository:  svcCtx.ArticleRepository,
		identityRepository: svcCtx.IdentityRepository,
	}
}

// 删除文章
func (l *DeleteArticleLogic) DeleteArticle(in *articlerpc.DeleteArticleRequest) (*articlerpc.Response, error) {
	// todo: add your logic here and delete this line
	// 删除文章 需要身份校验
	deleterID := l.ctxdata.GetUserIDFromContext(l.ctx)
	if deleterID == "" {
		l.logger.Error("article/deletearticlelogic：用户ID为空，疑似绕过api层进行请求（暂未锁定IP）")
		return l.articleResponse.RPCFailResponse("用户ID为空，请正常访问资源！", errcode.Unauthorized)
	}

	// 校验发起删除请求者的用户权限 至少管理员能操作
	deleterRole, err := l.identityRepository.GetUserRole(deleterID)
	if deleterRole < 1 { // 1管理员 2超级管理员
		l.logger.Error("article/deletearticle：普通用户访问删除文章操作，需要严重标记该用户", zap.String("deleterID", deleterID), zap.Int("role", deleterRole))
		return l.articleResponse.RPCFailResponse("权限不足", errcode.Forbidden)
	}

	// 校验文章ID
	articleIDOK, err := l.commonUtils.CheckArticleID(in.ArticleID)
	if err != nil {
		l.logger.Error("article/deletearticle：校验文章ID失败", zap.String("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.RPCFailResponseWithErr(err)
	}
	if !articleIDOK {
		l.logger.Error("article/deletearticle：校验文章ID错误，不存在的文章ID", zap.String("ArticleID", in.ArticleID))
		return l.articleResponse.RPCFailResponse("文章不存在！", errcode.NotFound)
	}

	// 根据文章ID获取文章发布者（此操作也可以侧面测出文章ID是否存在） 并得到文章发布者的身份权限
	publisherID, err := l.articleRepository.GetPublisherIDByArticleID(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("article/deletearticle：数据库根据文章ID查询发布者ID操作失败", zap.String("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.RPCFailResponse("数据库操作失败！", errcode.ServerError)
	}
	if publisherID == "" {
		l.logger.Error("article/deletearticle：数据库根据文章ID查询到的发布者ID为空", zap.String("Article", in.ArticleID))
		return l.articleResponse.RPCFailResponse("要删除的文章不存在！", errcode.NotFound)
	}

	// 判断删除者与发布者是否是同一个
	if deleterID != publisherID { // 不是同一个则对比权限
		// 获取发布者的身份权限
		publisherRole, err := l.identityRepository.GetUserRole(publisherID)
		if err != nil {
			l.logger.Error("article/deletearticle：数据库查询发布者身份权限失败", zap.String("PublisherID", publisherID), zap.String("ERROR", err.Error()))
			return l.articleResponse.RPCFailResponse("数据库操作失败！", errcode.ServerError)
		}
		if deleterRole <= publisherRole { // 相同权限不可删除 只有高一级才可执行
			l.logger.Error("article/deletearticle：删除操作者的身份权限不高于发布者权限，删除操作失败，请求已拦截",
				zap.String("PublisherID", publisherID), zap.Int("PublisherRole", publisherRole),
				zap.String("DeleterID", deleterID), zap.Int("DeleterRole", deleterRole))
			return l.articleResponse.RPCFailResponse("权限不高于发布者，删除失败！", errcode.Forbidden)
		}
		// 过滤得到的是 删除者的权限大于发布者
	}
	// 执行删除操作
	deleteOK, err := l.articleRepository.DeleteArticle(l.ctx, in.ArticleID, deleterID)
	if err != nil {
		l.logger.Error("article/deletearticle：数据库删除文章失败", zap.String("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.RPCFailResponse("数据库操作失败！", errcode.ServerError)
	}
	if !deleteOK {
		l.logger.Error("article/deletearticle：数据库删除文章数量为0，请查库检查", zap.String("ArticleID", in.ArticleID))
	}

	l.logger.Info("article/deletearticle：文章删除成功", zap.String("ArticleID", in.ArticleID), zap.String("DeleterID", deleterID))
	return l.articleResponse.RPCSuccessResponse("文章删除成功！")
}
