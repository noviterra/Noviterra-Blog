// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"Noviterra/server/app/article/articleapi/internal/svc"
	"Noviterra/server/app/article/articleapi/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLikesArticleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查看用户点赞文章列表
func NewGetUserLikesArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLikesArticleListLogic {
	return &GetUserLikesArticleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLikesArticleListLogic) GetUserLikesArticleList(req *types.GetUserLikesListRequest) (resp *types.GetUserLikesListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
