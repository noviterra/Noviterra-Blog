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
	getListArticleLogger interface {
		Error(msg string, fields ...zap.Field)
		Info(msg string, fields ...zap.Field)
	}
	getListArticleValidate interface {
		ValidatePage(page, pageSize int32) (int, int, error)
	}
	getListArticleResponse interface {
		SuccessGetArticleListResponse(resp *articlerpc.GetArticleListResponse,
			msg string, articles []model.ArticleList, totalCount int64) (*articlerpc.GetArticleListResponse, error)
		FailGetArticleListResponse(resp *articlerpc.GetArticleListResponse, msg string,
			status int32, code int64) (*articlerpc.GetArticleListResponse, error)
		FailGetArticleListResponseWithErr(resp *articlerpc.GetArticleListResponse, err error) (*articlerpc.GetArticleListResponse, error)
	}
	getListArticleRepository interface {
		GetArticleList(ctx context.Context, page, pageSize int) ([]model.ArticleList, int64, error)
	}
)

type GetArticleListLogic struct {
	ctx context.Context

	logger          getListArticleLogger
	articleValidate getListArticleValidate
	articleResponse getListArticleResponse
	repository      getListArticleRepository
}

func NewGetArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleListLogic {
	return &GetArticleListLogic{
		ctx:             ctx,
		logger:          svcCtx.Logger,
		articleValidate: svcCtx.ArticleValidate,
		articleResponse: svcCtx.ArticleResponse,
		repository:      svcCtx.ArticleRepository,
	}
}

// 查询文章列表
func (l *GetArticleListLogic) GetArticleList(in *articlerpc.GetArticleListRequest) (*articlerpc.GetArticleListResponse, error) {
	// todo: add your logic here and delete this line
	// 拦截器会将非法绕进rpc层的请求 不需要进行用户身份校验了
	resp := new(articlerpc.GetArticleListResponse)
	// 参数校验
	page, pageSize, err := l.articleValidate.ValidatePage(in.Page, in.PageSize)
	if err != nil { // 校验函数返回的为自定义err
		l.logger.Error("请求参数校验失败：page/page-size",
			zap.Int32("page", in.Page),
			zap.Int32("page-size", in.PageSize),
			zap.Error(err),
		)
		return l.articleResponse.FailGetArticleListResponseWithErr(resp, err)
	}
	if page == 0 || pageSize == 0 {
		l.logger.Error("请求参数校验失败：page/page-size",
			zap.Int32("page", in.Page),
			zap.Int32("page-size", in.PageSize),
		)
		return l.articleResponse.FailGetArticleListResponse(resp, "页码规模异常！",
			errcode.InvalidParam, errcode.ValidateArticleRequestParamFail)
	}

	// 查库
	articleList, totalCount, err := l.repository.GetArticleList(l.ctx, page, pageSize)
	if err != nil {
		l.logger.Error("数据库操作失败：分页查询文章列表失败",
			zap.Int("page", page),
			zap.Int("page-size", pageSize),
			zap.Error(err),
		)
		return l.articleResponse.FailGetArticleListResponse(resp, "数据库查询失败！",
			errcode.InternalError, errcode.GetArticleListFail)
	}
	if totalCount == 0 {
		l.logger.Info("分页查询文章列表结果为空",
			zap.Int("page", page),
			zap.Int("page-size", pageSize),
			zap.Int64("article-count", totalCount),
		)
	}

	l.logger.Info("分页查询文章列表成功",
		zap.Int("page", page),
		zap.Int("page-size", pageSize),
		zap.Int64("article-count", totalCount),
	)
	return l.articleResponse.SuccessGetArticleListResponse(resp, "列表加载完成！", articleList, totalCount)
}
