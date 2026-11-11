package logic

import (
	"Noviterra/server/utils/errcode"
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"go.uber.org/zap"
)

type (
	updateArticleCoverValidate interface {
		ValidateSnowID(id int64) (int64, error)
		CheckImageFile(fileName, fileDir string) (bool, error)
		CheckArticleCover(cover string, articleID int64) (bool, error)
	}
	updateArticleCoverCtxData interface {
		GetUserIDFromContext(ctx context.Context) string
	}
	updateArticleCoverLogger interface {
		Error(msg string, field ...zap.Field)
		Info(msg string, field ...zap.Field)
	}
	updateArticleCoverIdentityRepository interface {
		GetUserRole(ctx context.Context, articleID string) (int, error)
	}
	updateCoverArticleRepository interface {
		GetPublisherIDByArticleID(ctx context.Context, articleID int64) (string, error)
		UpdateArticleCover(ctx context.Context, articleID int64, cover string) (bool, error)
	}
	updateCoverArticleResponse interface {
		SuccessUpdateArticleCoverResponse(resp *articlerpc.UpdateArticleCoverResponse,
			msg, cover string) (*articlerpc.UpdateArticleCoverResponse, error)
		FailUpdateArticleCoverResponse(resp *articlerpc.UpdateArticleCoverResponse,
			msg string, status int32, code int64) (*articlerpc.UpdateArticleCoverResponse, error)
		FailUpdateArticleCoverResponseWithErr(resp *articlerpc.UpdateArticleCoverResponse,
			err error) (*articlerpc.UpdateArticleCoverResponse, error)
	}
)

type UpdateArticleCoverLogic struct {
	ctx                context.Context
	articleValidate    updateArticleCoverValidate
	ctxdata            updateArticleCoverCtxData
	logger             updateArticleCoverLogger
	identityRepository updateArticleCoverIdentityRepository
	articleRepository  updateCoverArticleRepository
	articleResponse    updateCoverArticleResponse
}

func NewUpdateArticleCoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateArticleCoverLogic {
	return &UpdateArticleCoverLogic{
		ctx:                ctx,
		articleValidate:    svcCtx.ArticleValidate,
		ctxdata:            svcCtx.CtxData,
		logger:             svcCtx.Logger,
		identityRepository: svcCtx.IdentityRepository,
		articleRepository:  svcCtx.ArticleRepository,
		articleResponse:    svcCtx.ArticleResponse,
	}
}

// 修改文章封面
func (l *UpdateArticleCoverLogic) UpdateArticleCover(in *articlerpc.UpdateArticleCoverRequest) (*articlerpc.UpdateArticleCoverResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(articlerpc.UpdateArticleCoverResponse)
	// 用户信息校验
	updaterID := l.ctxdata.GetUserIDFromContext(l.ctx)
	if updaterID == "" {
		l.logger.Error("上下文获取操作用户ID失败")
	}

	// 校验文章ID
	articleID, err := l.articleValidate.ValidateSnowID(in.ArticleID)
	if err != nil {
		l.logger.Error("校验文章ID参数失败",
			zap.Int64("article-id", in.ArticleID),
			zap.Error(err),
		)
		return l.articleResponse.FailUpdateArticleCoverResponseWithErr(resp, err)
	}
	if articleID == 0 {
		l.logger.Error("校验文章ID失败：文章ID为空",
			zap.Int64("article-id", in.ArticleID),
		)
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "信息校验失败！",
			errcode.InvalidParam, errcode.ValidateArticleIDFail)
	}

	// 联合文章ID校验传入的cover字段
	coverOK, err := l.articleValidate.CheckArticleCover(in.NewCover, in.ArticleID)
	if err != nil {
		l.logger.Error("校验文章封面参数失败",
			zap.Int64("article-id", articleID),
			zap.String("updater-info", updaterID),
			zap.String("cover", in.NewCover),
			zap.Error(err),
		)
		return l.articleResponse.FailUpdateArticleCoverResponseWithErr(resp, err)
	}
	if !coverOK {
		l.logger.Error("校验文章封面失败：封面为空",
			zap.Int64("article-id", articleID),
			zap.String("updater-info", updaterID),
			zap.String("cover", in.NewCover),
		)
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "文章封面为空！",
			errcode.InvalidParam, errcode.ValidateArticleCoverFail)
	}

	// 获取用户权限
	updaterRole, err := l.identityRepository.GetUserRole(l.ctx, updaterID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据用户ID查询用户身份权限等级失败",
			zap.Int64("article-id", articleID),
			zap.String("updater-info", updaterID),
			zap.Error(err),
		)
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.GetUserInfoFail)
	}
	if updaterRole < 1 {
		l.logger.Error("用户权限不足，无法修改文章封面",
			zap.Int64("article-id", articleID),
			zap.String("updater-info", updaterID),
			zap.Int("updater-role", updaterRole),
		)
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "权限不足！",
			errcode.Forbidden, errcode.ArticleOperationOutOfBounds)
	}

	// 根据文章ID得到发布者ID
	publisherID, err := l.articleRepository.GetPublisherIDByArticleID(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据文章ID查询发布者ID失败",
			zap.Int64("article-id", articleID),
			zap.Error(err),
		)
		return l.articleResponse.FailUpdateArticleCoverResponseWithErr(resp, err)
	}
	if publisherID == "" {
		l.logger.Error("根据文章ID查询到的发布者ID为空",
			zap.Int64("article-id", articleID),
		)
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "信息校验",
			errcode.InvalidParam, errcode.ArticleNotFound)
	}

	// 更新者与发布者需要是同一个
	if updaterID != publisherID {
		l.logger.Error("非文章发布者尝试修改文章封面，操作被拦截",
			zap.Int64("article-id", articleID),
			zap.String("updater-info", updaterID),
			zap.String("publisher-info", publisherID),
		)
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "信息校验失败！",
			errcode.Forbidden, errcode.ArticleOperationOutOfBounds)
	}

	// 数据库更新
	updateOPOK, err := l.articleRepository.UpdateArticleCover(l.ctx, in.ArticleID, in.NewCover)
	if err != nil {
		l.logger.Error("数据库操作失败：更新文章封面失败",
			zap.Int64("article-id", articleID),
			zap.String("updater-info", updaterID),
			zap.String("cover", in.NewCover),
			zap.Error(err),
		)
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "系统繁忙！",
			errcode.InternalError, errcode.UpdateArticleCoverFail)
	}
	if !updateOPOK {
		l.logger.Error("数据库更新文章封面影响行数为0，更新未生效",
			zap.Int64("article-id", articleID),
			zap.String("updater-info", updaterID),
			zap.String("cover", in.NewCover),
		)
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "更新失败！",
			errcode.InvalidParam, errcode.UpdateArticleCoverFail)
	}

	// 返回一个路径
	l.logger.Info("修改文章封面成功",
		zap.Int64("article-id", articleID),
		zap.String("updater-info", updaterID),
		zap.String("cover", in.NewCover),
	)
	return l.articleResponse.SuccessUpdateArticleCoverResponse(resp, "封面更新完成！",
		"/static/articlecover/"+in.NewCover)
}
