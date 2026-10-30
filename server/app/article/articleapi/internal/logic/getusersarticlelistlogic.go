// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"Noviterra/server/app/article/articleapi/internal/svc"
	"Noviterra/server/app/article/articleapi/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUsersArticleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查看用户发布文章列表
func NewGetUsersArticleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUsersArticleListLogic {
	return &GetUsersArticleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUsersArticleListLogic) GetUsersArticleList(req *types.GetUsersArticleListRequest) (resp *types.GetUsersArticleListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
