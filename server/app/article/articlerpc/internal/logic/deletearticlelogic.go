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
		GetUserRole(ctx context.Context, id string) (int, error)
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
		l.logger.Error("上下文获取操作用户ID失败")
		return l.articleResponse.FailResponse(resp, "用户ID为空，请正常访问资源！",
			errcode.Unauthorized, errcode.GetUserInfoFail)
	}

	// 校验发起删除请求者的用户权限 至少管理员能操作
	deleterRole, err := l.identityRepository.GetUserRole(l.ctx, deleterID)
	if deleterRole < 1 { // 1管理员 2超级管理员
		l.logger.Error("操作者权限不足或查询用户身份权限等级失败",
			zap.String("operator-info", deleterID),
			zap.Int("operator-role", deleterRole),
			zap.Error(err),
		)
		return l.articleResponse.FailResponse(resp, "权限不足",
			errcode.Forbidden, errcode.ArticleOperationOutOfBounds)
	}

	// 校验文章ID
	articleID, err := l.articleValidate.ValidateSnowID(in.ArticleID)
	if err != nil {
		l.logger.Error("校验文章ID参数失败",
			zap.Int64("article-id", in.ArticleID),
			zap.Error(err),
		)
		return l.articleResponse.FailResponseWithErr(resp, err)
	}
	if articleID == 0 {
		l.logger.Error("校验文章ID失败：文章ID为空",
			zap.Int64("article-id", in.ArticleID),
		)
		return l.articleResponse.FailResponse(resp, "文章不存在！",
			errcode.NotFound, errcode.ValidateArticleIDFail)
	}

	// 根据文章ID获取文章发布者（此操作也可以侧面测出文章ID是否存在） 并得到文章发布者的身份权限
	publisherID, err := l.articleRepository.GetPublisherIDByArticleID(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据文章ID查询发布者ID失败",
			zap.Int64("article-id", in.ArticleID),
			zap.Error(err),
		)
		return l.articleResponse.FailResponse(resp, "数据库操作失败！",
			errcode.InternalError, errcode.GetArticleInfoFail)
	}
	if publisherID == "" {
		l.logger.Error("根据文章ID查询到的发布者ID为空",
			zap.Int64("article-id", in.ArticleID),
		)
		return l.articleResponse.FailResponse(resp, "要删除的文章不存在！",
			errcode.NotFound, errcode.ArticleNotFound)
	}

	// 判断删除者与发布者是否是同一个
	if deleterID != publisherID { // 不是同一个则对比权限
		// 获取发布者的身份权限
		publisherRole, err := l.identityRepository.GetUserRole(l.ctx, publisherID)
		if err != nil {
			l.logger.Error("数据库操作失败：根据发布者ID查询用户身份权限等级失败",
				zap.String("publisher-info", publisherID),
				zap.Error(err),
			)
			return l.articleResponse.FailResponse(resp, "数据库操作失败！",
				errcode.InternalError, errcode.GetUserInfoFail)
		}
		if deleterRole <= publisherRole { // 相同权限不可删除 只有高一级才可执行
			l.logger.Error("操作者权限不高于发布者，删除操作被拦截",
				zap.String("operator-info", deleterID),
				zap.Int("operator-role", deleterRole),
				zap.String("publisher-info", publisherID),
				zap.Int("publisher-role", publisherRole),
			)
			return l.articleResponse.FailResponse(resp, "权限不高于发布者，删除失败！",
				errcode.Forbidden, errcode.ArticleOperationOutOfBounds)
		}
		// 过滤得到的是 删除者的权限大于发布者
	}
	// 执行删除操作
	deleteOK, err := l.articleRepository.DeleteArticle(l.ctx, in.ArticleID, deleterID)
	if err != nil {
		l.logger.Error("数据库操作失败：删除文章失败",
			zap.Int64("article-id", in.ArticleID),
			zap.String("operator-info", deleterID),
			zap.Error(err),
		)
		return l.articleResponse.FailResponse(resp, "数据库操作失败！",
			errcode.InternalError, errcode.DeleteArticleFail)
	}
	if !deleteOK {
		l.logger.Error("数据库删除文章影响行数为0，删除未生效",
			zap.Int64("article-id", in.ArticleID),
		)
	}

	l.logger.Info("删除文章成功",
		zap.Int64("article-id", in.ArticleID),
		zap.String("operator-info", deleterID),
		zap.Int("operator-role", deleterRole),
	)
	return l.articleResponse.SuccessResponse(resp, "文章删除成功！")
}
