package logic

import (
	"context"

	"Noviterra/server/app/article/articlerpc/internal/svc"
	"Noviterra/server/app/article/articlerpc/types/articlerpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetArticleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleListLogic {
	return &GetArticleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询文章列表
func (l *GetArticleListLogic) GetArticleList(in *articlerpc.GetArticleListRequest) (*articlerpc.GetArticleListResponse, error) {
	// todo: add your logic here and delete this line

	return &articlerpc.GetArticleListResponse{}, nil
}
