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

type getArchiveArticleRepository interface {
	GetArchiveList(ctx context.Context) ([]model.ArticleArchive, error)
}

type getArchiveLogger interface {
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
}

type getArchiveArticleResponse interface {
	RPCSuccessGetArchiveResponse(resp *articlerpc.GetArchiveResponse, archives []model.ArticleArchive) (*articlerpc.GetArchiveResponse, error)
	RPCFailGetArchiveResponse(resp *articlerpc.GetArchiveResponse, msg string, code int64) (*articlerpc.GetArchiveResponse, error)
	RPCFailGetArchiveResponseWithErr(resp *articlerpc.GetArchiveResponse, err error) (*articlerpc.GetArchiveResponse, error)
}

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
	// 缓存

	// 查库 进行列表映射
	list, err := l.articleRepository.GetArchiveList(l.ctx)
	if err != nil {
		l.logger.Error("article/getArchive：数据库查询得到归档列表失败",
			zap.String("ERROR", err.Error()))
		return l.articleResponse.RPCFailGetArchiveResponse(resp, "数据库操作失败！", errcode.InternalError)
	}
	if len(list) == 0 {
		l.logger.Info("article/getArchive：归档数据为空",
			zap.Int64("Time", time.Now().Unix()))
	}

	// 回写缓存

	return l.articleResponse.RPCSuccessGetArchiveResponse(resp, list)
}
