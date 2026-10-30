// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"Noviterra/server/app/article/articleapi/internal/svc"
	"Noviterra/server/app/article/articleapi/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserCollectArticleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查看用户收藏文章列表
func NewGetUserCollectArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserCollectArticleListLogic {
	return &GetUserCollectArticleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserCollectArticleListLogic) GetUserCollectArticleList(req *types.GetUserCollectListRequest) (resp *types.GetUserCollectListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
