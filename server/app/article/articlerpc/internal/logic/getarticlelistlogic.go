package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"
	"fmt"
	"time"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"go.uber.org/zap"
)

type getListArticleLogger interface {
	Error(msg string, fields ...zap.Field)
}

type getListArticleCommonUntils interface {
	ValidatePage(page, pageSize int32) (int, int, error)
}

type getListArticleResponse interface {
	RPCSuccessGetArticleListResponse(msg string, articles []*articlerpc.ArticleList, totalCount int32) (*articlerpc.GetArticleListResponse, error)
	RPCFailGetArticleListResponse(msg string, code int64) (*articlerpc.GetArticleListResponse, error)
	RPCFailGetArticleListResponseWithErr(err error) (*articlerpc.GetArticleListResponse, error)
}

type getListArticleRepository interface {
	GetArticleList(ctx context.Context, page, pageSize int) ([]model.ArticleList, int64, error)
}

type GetArticleListLogic struct {
	ctx context.Context

	logger          getListArticleLogger
	commonUtils     getListArticleCommonUntils
	articleResponse getListArticleResponse
	repository      getListArticleRepository
}

func NewGetArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleListLogic {
	return &GetArticleListLogic{
		ctx:             ctx,
		logger:          svcCtx.Logger,
		commonUtils:     svcCtx.CommonUtils,
		articleResponse: svcCtx.ArticleResponse,
		repository:      svcCtx.ArticleRepository,
	}
}

// 查询文章列表
func (l *GetArticleListLogic) GetArticleList(in *articlerpc.GetArticleListRequest) (*articlerpc.GetArticleListResponse, error) {
	// todo: add your logic here and delete this line
	// 拦截器会将非法绕进rpc层的请求 不需要进行用户身份校验了
	// 参数校验
	page, pageSize, err := l.commonUtils.ValidatePage(in.Page, in.PageSize)
	if err != nil { // 校验函数返回的为自定义err
		l.logger.Error("article/getArticleList：page/pagesize参数校验失败", zap.String("ERROR", err.Error()))
		return l.articleResponse.RPCFailGetArticleListResponseWithErr(err)
	}
	if page == 0 || pageSize == 0 {
		l.logger.Error("article/getArticleList：page/pagesize参数校验失败")
		return l.articleResponse.RPCFailGetArticleListResponse("页码规模异常！", errcode.BadPageData)
	}

	// 查库
	articleList, totalCount, err := l.repository.GetArticleList(l.ctx, page, pageSize)
	if err != nil {
		l.logger.Error("article/getArticleList：数据库分页查询文章列表失败", zap.String("ERROR", err.Error()))
		return l.articleResponse.RPCFailGetArticleListResponse("数据库查询失败！", errcode.ServerError)
	}
	if totalCount == 0 {
		l.logger.Error("article/getArticleList：数据库查询结束", zap.Int64("ArticleCount", totalCount), zap.Time("CurrentTime", time.Now()))
		return l.articleResponse.RPCSuccessGetArticleListResponse("当前暂未有公布的文章", nil, int32(totalCount))
	}

	// 继续数据映射
	resultList := make([]*articlerpc.ArticleList, 0, totalCount)
	for _, list := range articleList {
		ls := new(articlerpc.ArticleList)

		ls.ArticleID = list.ArticleID
		ls.Title = list.Title
		ls.Information = list.Information
		ls.Tags = list.Tags
		ls.Cover = list.Cover
		ls.CreatedAt = list.CreatedAt.Unix()

		resultList = append(resultList, ls)
	}

	return l.articleResponse.RPCSuccessGetArticleListResponse(fmt.Sprintf("查询到共计%d篇文章", totalCount), resultList, int32(totalCount))
}
