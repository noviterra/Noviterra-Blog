package logic

import (
	"context"

	"Noviterra/server/app/comment/commentrpc/internal/svc"
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTopCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTopCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTopCommentLogic {
	return &GetTopCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 得到某个位置的全部顶层评论
func (l *GetTopCommentLogic) GetTopComment(in *commentrpc.GetTopCommentRequest) (*commentrpc.GetTopCommentResponse, error) {
	// todo: add your logic here and delete this line

	return &commentrpc.GetTopCommentResponse{}, nil
}
