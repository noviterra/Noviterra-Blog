package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"

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
		GetUserRole(ctx context.Context, userID string) (int, error)
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
		l.logger.Error("上下文获取操作用户ID失败")
		return l.articleResponse.FailResponse(resp, "用户信息异常！",
			errcode.Unauthorized, errcode.GetUserInfoFail)
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
		return l.articleResponse.FailResponse(resp, "请求的文章不存在或已经删除！",
			errcode.NotFound, errcode.ValidateArticleIDFail)
	}

	// 获取用户权限
	updaterRole, err := l.identityRepository.GetUserRole(l.ctx, updaterID)
	if updaterRole < 1 {
		l.logger.Error("操作者权限不足或查询用户身份权限等级失败",
			zap.Int64("article-id", articleID),
			zap.String("updater-info", updaterID),
			zap.Int("updater-role", updaterRole),
			zap.Error(err),
		)
		return l.articleResponse.FailResponse(resp, "用户信息异常！",
			errcode.Forbidden, errcode.ArticleOperationOutOfBounds)
	}

	// 根据文章ID得到发布者ID
	publisherID, err := l.articleRepository.GetPublisherIDByArticleID(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据文章ID查询发布者ID失败",
			zap.Int64("article-id", articleID),
			zap.Error(err),
		)
		return l.articleResponse.FailResponse(resp, "数据库操作失败！",
			errcode.InternalError, errcode.GetArticleInfoFail)
	}
	if publisherID == "" {
		l.logger.Error("根据文章ID查询到的发布者ID为空",
			zap.Int64("article-id", articleID),
		)
		return l.articleResponse.FailResponse(resp, "需要更新的文章不存在！",
			errcode.NotFound, errcode.ArticleNotFound)
	}

	// 更新者与发布者需要是同一个
	if updaterID != publisherID {
		l.logger.Error("非文章发布者尝试修改文章，操作被拦截",
			zap.Int64("article-id", articleID),
			zap.String("updater-info", updaterID),
			zap.String("publisher-info", publisherID),
		)
		return l.articleResponse.FailResponse(resp, "无法修改非本人发布的文章！",
			errcode.Forbidden, errcode.ArticleOperationOutOfBounds)
	}

	// 更新操作
	updateOPOK, err := l.articleRepository.UpdateArticleInfo(l.ctx, in.ArticleID, &model.UpdateArticle{
		Title:       in.Title,
		Information: in.Information,
		Content:     in.Content,
		Tags:        in.Tags,
	})
	if err != nil {
		l.logger.Error("数据库操作失败：更新文章信息失败",
			zap.Int64("article-id", articleID),
			zap.String("updater-info", updaterID),
			zap.Error(err),
		)
		return l.articleResponse.FailResponse(resp, "数据库异常！",
			errcode.InternalError, errcode.UpdateArticleFail)
	}
	if !updateOPOK {
		l.logger.Error("数据库更新文章信息影响行数为0，更新未生效",
			zap.Int64("article-id", articleID),
			zap.String("updater-info", updaterID),
		)
	}

	l.logger.Info("更新文章信息成功",
		zap.Int64("article-id", articleID),
		zap.String("updater-info", updaterID),
		zap.Int("updater-role", updaterRole),
	)
	return l.articleResponse.SuccessResponse(resp, "更新文章操作完成！")
}
