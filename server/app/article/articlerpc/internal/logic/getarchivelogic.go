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
	getArchiveArticleRepository interface {
		GetArchiveList(ctx context.Context) ([]model.ArticleArchive, error)
	}
	getArchiveLogger interface {
		Error(msg string, fields ...zap.Field)
		Info(msg string, fields ...zap.Field)
		Warn(msg string, fields ...zap.Field)
	}
	getArchiveArticleResponse interface {
		SuccessGetArchiveResponse(resp *articlerpc.GetArchiveResponse, msg string, archives []model.ArticleArchive) (*articlerpc.GetArchiveResponse, error)
		FailGetArchiveResponse(resp *articlerpc.GetArchiveResponse, msg string, status int32, code int64) (*articlerpc.GetArchiveResponse, error)
		FailGetArchiveResponseWithErr(resp *articlerpc.GetArchiveResponse, err error) (*articlerpc.GetArchiveResponse, error)
	}
)

type GetArchiveLogic struct {
	ctx               context.Context
	articleRepository getArchiveArticleRepository
	logger            getArchiveLogger
	articleResponse   getArchiveArticleResponse
}

func NewGetArchiveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArchiveLogic {
	return &GetArchiveLogic{
		ctx:               ctx,
		articleRepository: svcCtx.ArticleRepository,
		logger:            svcCtx.Logger,
		articleResponse:   svcCtx.ArticleResponse,
	}
}

// 得到文章归档
func (l *GetArchiveLogic) GetArchive(in *articlerpc.GetArchiveRequest) (*articlerpc.GetArchiveResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(articlerpc.GetArchiveResponse)
	// 请求体为空 不需要校验参数

	// 查库 进行列表映射
	list, err := l.articleRepository.GetArchiveList(l.ctx)
	if err != nil {
		l.logger.Error("数据库操作失败：查询文章归档列表失败",
			zap.Error(err),
		)
		return l.articleResponse.FailGetArchiveResponse(resp, "数据库操作失败！",
			errcode.InternalError, errcode.GetArticleArchiveFail)
	}
	if len(list) == 0 {
		l.logger.Info("文章归档列表为空",
			zap.Int("archive-count", len(list)),
		)
	}

	l.logger.Info("获取文章归档列表成功",
		zap.Int("archive-count", len(list)),
	)
	return l.articleResponse.SuccessGetArchiveResponse(resp, "列表加载完成！ ", list)
}
