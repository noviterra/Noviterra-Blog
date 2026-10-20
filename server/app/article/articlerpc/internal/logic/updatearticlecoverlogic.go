package logic

import (
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"go.uber.org/zap"
)

type updateArticleCoverCommonUtils interface {
	CheckImageFile(fileName, fileDir string) (bool, error)
	CheckArticleID(articleID string) (bool, error)
	CheckArticleCover(cover, articleID string) (bool, error)
}

type updateArticleCoverCtxData interface {
	GetUserID(ctx context.Context) (string, bool)
}

type updateArticleCoverLogger interface {
	Error(msg string, field ...zap.Field)
	Info(msg string, field ...zap.Field)
}

type updateArticleCoverIdentityRepository interface {
	GetUserRole(articleID string) (int, error)
}

type updateCoverArticleRepository interface {
	GetPublisherIDByArticleID(ctx context.Context, articleID string) (string, error)
	UpdateArticleCover(ctx context.Context, articleID, cover string) (bool, error)
}

type updateCoverArticleResponse interface {
	RPCSuccessUpdateArticleCoverResponse(msg, cover string) (*articlerpc.UpdateArticleCoverResponse, error)
	RPCFailUpdateArticleCoverResponse(msg string, code int64) (*articlerpc.UpdateArticleCoverResponse, error)
	RPCFailUpdateArticleCoverResponseWithErr(error) (*articlerpc.UpdateArticleCoverResponse, error)
}

type UpdateArticleCoverLogic struct {
	ctx                context.Context
	commonUtils        updateArticleCoverCommonUtils
	ctxdata            updateArticleCoverCtxData
	logger             updateArticleCoverLogger
	identityRepository updateArticleCoverIdentityRepository
	articleRepository  updateCoverArticleRepository
	articleResponse    updateCoverArticleResponse
}

func NewUpdateArticleCoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateArticleCoverLogic {
	return &UpdateArticleCoverLogic{
		ctx: ctx,

		commonUtils:        svcCtx.CommonUtils,
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
	// 用户信息校验
	updaterID, ok := l.ctxdata.GetUserID(l.ctx)
	if !ok {
		l.logger.Error("article/updatearticlecover：上下文获取用户ID失败")
	}
	if updaterID == "" {
		l.logger.Error("article/updatearticlecover：上下文获取用户ID为空")
	}

	// 校验文章ID
	articleIDOK, err := l.commonUtils.CheckArticleID(in.ArticleID)
	if err != nil {
		l.logger.Error("article/updatearticle：请求中的文章ID格式有误",
			zap.String("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
	}
	if !articleIDOK {
		l.logger.Error("article/updatearticle：请求中的文章ID格式有误",
			zap.String("ArticleID", in.ArticleID))
	}

	// 联合文章ID校验传入的cover字段
	coverOK, err := l.commonUtils.CheckArticleCover(in.NewCover, in.ArticleID)
	if err != nil {
		l.logger.Error("article/updatearticlecover：校验文章封面结构失败", zap.String("ArticleID", in.ArticleID),
			zap.String("updaterID", updaterID), zap.String("ERROR", err.Error()), zap.String("cover", in.NewCover))
	}
	if !coverOK {
		l.logger.Error("article/updatearticlecover：校验文章封面结构失败", zap.String("ArticleID", in.ArticleID),
			zap.String("updaterID", updaterID), zap.String("cover", in.NewCover))
	}

	// 获取用户权限
	updaterRole, err := l.identityRepository.GetUserRole(updaterID)
	if err != nil {
		l.logger.Error("article/updatearticlecover：数据库获取用户权限失败", zap.String("ArticleID", in.ArticleID),
			zap.String("updaterID", updaterID), zap.String("ERROR", err.Error()))
	}
	if updaterRole < 1 {
		l.logger.Error("article/updatearticlecover：普通用户访问修改文章封面操作", zap.String("ArticleID", in.ArticleID),
			zap.String("UpdaterID", updaterID), zap.Int("UpdaterRole", updaterRole))
	}

	// 根据文章ID得到发布者ID
	publisherID, err := l.articleRepository.GetPublisherIDByArticleID(l.ctx, in.ArticleID)
	if err != nil {
		l.logger.Error("article/updatearticle：数据库根据文章ID查询发布者ID失败",
			zap.String("ArticleID", in.ArticleID), zap.String("ERROR", err.Error()))
	}
	if publisherID == "" {
		l.logger.Error("article/updatearticle：数据库根据文章ID查询的发布者ID为空",
			zap.String("ArticleID", in.ArticleID), zap.String("PublisherID", publisherID))
	}

	// 更新者与发布者需要是同一个
	if updaterID != publisherID {
		l.logger.Error("article/updatearticle：非发布者预更新此文章，已拦截",
			zap.String("ArticleID", in.ArticleID),
			zap.String("UpdaterID", updaterID), zap.String("PublisherID", publisherID))
	}

	// 数据库更新
	updateOPOK, err := l.articleRepository.UpdateArticleCover(l.ctx, in.ArticleID, in.NewCover)
	if err != nil {
		l.logger.Error("article/updatearticlecover：数据库更新文章封面失败", zap.String("ArticleID", in.ArticleID),
			zap.String("updaterID", updaterID), zap.String("ERROR", err.Error()), zap.String("cover", in.NewCover))
	}
	if !updateOPOK {
		l.logger.Error("article/updatearticlecover：数据库更新文章封面生效数量为0，更新失败", zap.String("ArticleID", in.ArticleID),
			zap.String("updaterID", updaterID), zap.String("cover", in.NewCover))
	}

	// 返回一个路径
	return l.articleResponse.RPCSuccessUpdateArticleCoverResponse("封面更新完成！", "/static/articlecover/"+in.NewCover)
}
