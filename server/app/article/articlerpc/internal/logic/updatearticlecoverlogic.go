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
		GetUserRole(articleID string) (int, error)
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
		l.logger.Error("article/updatearticlecover：上下文获取用户ID为空")
	}

	// 校验文章ID
	articleID, err := l.articleValidate.ValidateSnowID(in.ArticleID)
	if err != nil {
		l.logger.Error("article/updatearticle：请求中的文章ID格式有误",
			zap.Int64("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.FailUpdateArticleCoverResponseWithErr(resp, err)
	}
	if articleID == 0 {
		l.logger.Error("article/updatearticle：请求中的文章ID格式有误",
			zap.Int64("ArticleID", in.ArticleID))
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "信息校验失败！", errcode.InvalidParam, errcode.Failed)
	}

	// 联合文章ID校验传入的cover字段
	coverOK, err := l.articleValidate.CheckArticleCover(in.NewCover, in.ArticleID)
	if err != nil {
		l.logger.Error("article/updatearticlecover：校验文章封面结构失败", zap.Int64("ArticleID", in.ArticleID),
			zap.String("updaterID", updaterID), zap.String("ERROR", err.Error()), zap.String("cover", in.NewCover))
		return l.articleResponse.FailUpdateArticleCoverResponseWithErr(resp, err)
	}
	if !coverOK {
		l.logger.Error("article/updatearticlecover：校验文章封面结构失败", zap.Int64("ArticleID", in.ArticleID),
			zap.String("updaterID", updaterID), zap.String("cover", in.NewCover))
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "文章封面为空！", errcode.InvalidParam, errcode.Failed)
	}

	// 获取用户权限
	updaterRole, err := l.identityRepository.GetUserRole(updaterID)
	if err != nil {
		l.logger.Error("article/updatearticlecover：数据库获取用户权限失败", zap.Int64("ArticleID", in.ArticleID),
			zap.String("updaterID", updaterID), zap.String("ERROR", err.Error()))
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "系统繁忙，请稍后再试！", errcode.InternalError, errcode.Failed)
	}
	if updaterRole < 1 {
		l.logger.Error("article/updatearticlecover：普通用户访问修改文章封面操作", zap.Int64("ArticleID", in.ArticleID),
			zap.String("UpdaterID", updaterID), zap.Int("UpdaterRole", updaterRole))
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "权限不足！", errcode.Forbidden, errcode.Failed)
	}

	// 根据文章ID得到发布者ID
	publisherID, err := l.articleRepository.GetPublisherIDByArticleID(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("article/updatearticle：数据库根据文章ID查询发布者ID失败",
			zap.Int64("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
		return l.articleResponse.FailUpdateArticleCoverResponseWithErr(resp, err)
	}
	if publisherID == "" {
		l.logger.Error("article/updatearticle：数据库根据文章ID查询的发布者ID为空",
			zap.Int64("ArticleID", in.ArticleID), zap.String("PublisherID", publisherID))
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "信息校验", errcode.InvalidParam, errcode.Failed)
	}

	// 更新者与发布者需要是同一个
	if updaterID != publisherID {
		l.logger.Error("article/updatearticle：非发布者预更新此文章，已拦截",
			zap.Int64("ArticleID", in.ArticleID),
			zap.String("UpdaterID", updaterID), zap.String("PublisherID", publisherID))
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "信息校验失败！", errcode.Forbidden, errcode.Failed)
	}

	// 数据库更新
	updateOPOK, err := l.articleRepository.UpdateArticleCover(l.ctx, in.ArticleID, in.NewCover)
	if err != nil {
		l.logger.Error("article/updatearticlecover：数据库更新文章封面失败", zap.Int64("ArticleID", in.ArticleID),
			zap.String("updaterID", updaterID), zap.String("ERROR", err.Error()), zap.String("cover", in.NewCover))
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "系统繁忙！", errcode.InternalError, errcode.Failed)
	}
	if !updateOPOK {
		l.logger.Error("article/updatearticlecover：数据库更新文章封面生效数量为0，更新失败", zap.Int64("ArticleID", in.ArticleID),
			zap.String("updaterID", updaterID), zap.String("cover", in.NewCover))
		return l.articleResponse.FailUpdateArticleCoverResponse(resp, "更新失败！", errcode.InvalidParam, errcode.Failed)
	}

	// 返回一个路径
	return l.articleResponse.SuccessUpdateArticleCoverResponse(resp, "封面更新完成！", "/static/articlecover/"+in.NewCover)
}
