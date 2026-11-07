// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"Noviterra/server/app/comment/commentapi/internal/svc"
	"Noviterra/server/app/comment/commentapi/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTreeCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 展开某个评论
func NewGetTreeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTreeCommentLogic {
	return &GetTreeCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTreeCommentLogic) GetTreeComment(req *types.GetTreeCommentRequest) (resp *types.GetTreeCommentResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
