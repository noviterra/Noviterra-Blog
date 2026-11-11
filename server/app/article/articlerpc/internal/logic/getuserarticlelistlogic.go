package logic

import (
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"Noviterra/server/utils/validate"
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"go.uber.org/zap"
)

type (
	getUserArticleListArticleValidate interface {
		ValidateUserArticleList(in validate.InputValidateUserArticleList) (validate.OuputValidateUserArticleList, error)
	}
	getUserArticleListArticleResponse interface {
		SuccessGetArticleListResponse(resp *articlerpc.GetArticleListResponse,
			msg string, articles []model.ArticleList, totalCount int64) (*articlerpc.GetArticleListResponse, error)
		FailGetArticleListResponse(resp *articlerpc.GetArticleListResponse, msg string,
			status int32, code int64) (*articlerpc.GetArticleListResponse, error)
		FailGetArticleListResponseWithErr(resp *articlerpc.GetArticleListResponse, err error) (*articlerpc.GetArticleListResponse, error)
	}
	getUserArticleListIdentityRepository interface {
		GetUserRole(ctx context.Context, id string) (int, error)
	}
	getUserArticleListArticleRepository interface {
		GetUserPublishArticleList(ctx context.Context, page, pageSize int,
			userID string) ([]model.ArticleList, int64, error)
	}
	getUserArticleListLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type GetUserArticleListLogic struct {
	ctx                context.Context
	inentityValidate   getUserArticleListArticleValidate
	articleResponse    getUserArticleListArticleResponse
	identityRepository getUserArticleListIdentityRepository
	articleRepository  getUserArticleListArticleRepository
	logger             getUserArticleListLogger
}

func NewGetUserArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserArticleListLogic {
	return &GetUserArticleListLogic{
		ctx:                ctx,
		inentityValidate:   svcCtx.ArticleValidate,
		articleResponse:    svcCtx.ArticleResponse,
		identityRepository: svcCtx.IdentityRepository,
		articleRepository:  svcCtx.ArticleRepository,
		logger:             svcCtx.Logger,
	}
}

// 得到用户发布的全部文章列表
func (l *GetUserArticleListLogic) GetUserArticleList(in *articlerpc.GetUserArticleListRequest) (*articlerpc.GetArticleListResponse, error) {
	// todo: add your logic here and delete this line
	resp := new(articlerpc.GetArticleListResponse)
	// 参数校验
	data, err := l.inentityValidate.ValidateUserArticleList(validate.InputValidateUserArticleList{
		UserID:   in.UserID,
		Page:     in.Page,
		PageSize: in.PageSize,
	})
	if err != nil {
		l.logger.Error("请求参数校验失败：用户文章列表参数",
			zap.String("user-info", in.UserID),
			zap.Int32("page", in.Page),
			zap.Int32("page-size", in.PageSize),
			zap.Error(err),
		)
		return l.articleResponse.FailGetArticleListResponseWithErr(resp, err)
	}

	// 校验userID的权限
	// 只有role>=1 即管理员或超级管理员才能进行发布文章
	// 普通用户的该列表直接返回空就行
	userRole, err := l.identityRepository.GetUserRole(l.ctx, data.UserID)
	if err != nil {
		l.logger.Error("数据库操作失败：根据用户ID查询用户身份权限等级失败",
			zap.String("user-info", data.UserID),
			zap.Error(err),
		)
		return l.articleResponse.FailGetArticleListResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.GetUserInfoFail)
	}
	if userRole < 1 { // 查看的是普通用户
		l.logger.Info("用户身份权限为普通用户，返回空文章列表",
			zap.String("user-info", data.UserID),
			zap.Int("user-role", userRole),
		)
		return l.articleResponse.SuccessGetArticleListResponse(resp, "列表加载完成！",
			[]model.ArticleList{}, 0)
	}

	// 查询缓存

	// 查询用户发布的文章列表
	articleList, count, err := l.articleRepository.GetUserPublishArticleList(l.ctx, data.Page,
		data.PageSize, data.UserID)
	if err != nil {
		l.logger.Error("数据库操作失败：查询用户发布的文章列表失败",
			zap.String("user-info", data.UserID),
			zap.Int("page", data.Page),
			zap.Int("page-size", data.PageSize),
			zap.Error(err),
		)
		return l.articleResponse.FailGetArticleListResponse(resp, "系统繁忙，请稍后再试！",
			errcode.InternalError, errcode.GetArticleListFail)
	}

	// 回写缓存

	l.logger.Info("获取用户发布的文章列表成功",
		zap.String("user-info", data.UserID),
		zap.Int("page", data.Page),
		zap.Int("page-size", data.PageSize),
		zap.Int64("article-count", count),
	)
	return l.articleResponse.SuccessGetArticleListResponse(resp, "文章列表加载完成！", articleList, count)
}
