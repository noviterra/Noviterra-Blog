package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"
	"time"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"go.uber.org/zap"
)

type (
	updateArticleLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
		Warn(msg string, fields ...zap.Field)
	}

	updateArticleRepository interface {
		GetPublisherIDByArticleID(ctx context.Context, articleID int64) (string, error)
		UpdateArticleInfo(ctx context.Context, articleID int64, article *model.UpdateArticle) (bool, error)
	}

	updateArticleCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}

	updateArticleIdentityRepository interface {
		GetUserRole(userID string) (int, error)
	}
	updateArticleResponse interface {
		SuccessResponse(resp *articlerpc.Response, msg string) (*articlerpc.Response, error)
		FailResponse(resp *articlerpc.Response, msg string, status int32, code int64) (*articlerpc.Response, error)
		FailResponseWithErr(resp *articlerpc.Response, err error) (*articlerpc.Response, error)
	}
	updateArticleValidate interface {
		ValidateSnowID(id int64) (int64, error)
	}
)

type UpdateArticleLogic struct {
	ctx                context.Context
	logger             updateArticleLogger
	articleRepository  updateArticleRepository
	ctxdata            updateArticleCtxData
	articleResponse    updateArticleResponse
	articleValidate    updateArticleValidate
	identityRepository updateArticleIdentityRepository
}

func NewUpdateArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateArticleLogic {
	return &UpdateArticleLogic{
		ctx:                ctx,
		logger:             svcCtx.Logger,
		articleRepository:  svcCtx.ArticleRepository,
		ctxdata:            svcCtx.CtxData,
		articleResponse:    svcCtx.ArticleResponse,
		articleValidate:    svcCtx.ArticleValidate,
		identityRepository: svcCtx.IdentityRepository,
	}
}

// 修改文章
func (l *UpdateArticleLogic) UpdateArticle(in *articlerpc.UpdateArticleRequest) (*articlerpc.Response, error) {
	// todo: add your logic here and delete this line
	resp := new(articlerpc.Response)
	// 用户信息校验
	updaterID := l.ctxdata.GetUserIDFromContext(l.ctx)
	if updaterID == "" {
		l.logger.Error("article/updatearticle：上下文获取用户ID为空")
		return l.articleResponse.FailResponse(resp, "用户信息异常！", errcode.Unauthorized, errcode.Failed)
	}

	// 校验文章ID
	articleID, err := l.articleValidate.ValidateSnowID(in.ArticleID)
	if err != nil {
		l.logger.Error("article/updatearticle：请求中的文章ID格式有误",
			zap.Int64("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.FailResponseWithErr(resp, err)
	}
	if articleID == 0 {
		l.logger.Error("article/updatearticle：请求中的文章ID格式有误",
			zap.Int64("ArticleID", in.ArticleID))
		return l.articleResponse.FailResponse(resp, "请求的文章不存在或已经删除！", errcode.NotFound, errcode.Failed)
	}

	// 获取用户权限
	updaterRole, err := l.identityRepository.GetUserRole(updaterID)
	if updaterRole < 1 {
		l.logger.Error("article/updatearticle：普通用户访问修改文章操作",
			zap.String("UpdaterID", updaterID), zap.Int("UpdaterRole", updaterRole))
		return l.articleResponse.FailResponse(resp, "用户信息异常！", errcode.Forbidden, errcode.Failed)
	}

	// 根据文章ID得到发布者ID
	publisherID, err := l.articleRepository.GetPublisherIDByArticleID(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("article/updatearticle：数据库根据文章ID查询发布者ID失败",
			zap.Int64("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.FailResponse(resp, "数据库操作失败！", errcode.ServerError, errcode.Failed)
	}
	if publisherID == "" {
		l.logger.Error("article/updatearticle：数据库根据文章ID查询的发布者ID为空",
			zap.Int64("ArticleID", in.ArticleID), zap.String("PublisherID", publisherID))
		return l.articleResponse.FailResponse(resp, "需要更新的文章不存在！", errcode.NotFound, errcode.Failed)
	}

	// 更新者与发布者需要是同一个
	if updaterID != publisherID {
		l.logger.Error("article/updatearticle：非发布者预更新此文章，已拦截",
			zap.Int64("ArticleID", in.ArticleID),
			zap.String("UpdaterID", updaterID), zap.String("PublisherID", publisherID))
		return l.articleResponse.FailResponse(resp, "无法修改非本人发布的文章！", errcode.Forbidden, errcode.Failed)
	}

	// 更新操作
	updateOPOK, err := l.articleRepository.UpdateArticleInfo(l.ctx, in.ArticleID, &model.UpdateArticle{
		Title:       in.Title,
		Information: in.Information,
		Content:     in.Content,
		Tags:        in.Tags,
	})
	if err != nil {
		l.logger.Error("article/updatearticle：数据库更新文章字段失败",
			zap.Int64("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.FailResponse(resp, "数据库异常！", errcode.ServerError, errcode.Failed)
	}
	if !updateOPOK {
		l.logger.Warn("article/updatearticle：数据库更新文章字段未更新有效数据，更新数据条数为0")
	}

	l.logger.Info("article/updatearticle：更新文章操作完成",
		zap.Int64("ArticleID", in.ArticleID), zap.Time("UpdatedAt", time.Now()))
	return l.articleResponse.SuccessResponse(resp, "更新文章操作完成！")
}
