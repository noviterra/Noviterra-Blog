package logic

import (
	"context"

	"Noviterra/server/app/comment/commentrpc/internal/svc"
	"Noviterra/server/app/comment/commentrpc/types/commentrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTreeCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTreeCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTreeCommentLogic {
	return &GetTreeCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 得到某个评论的全部子评论
func (l *GetTreeCommentLogic) GetTreeComment(in *commentrpc.GetTreeCommentRequest) (*commentrpc.GetTopCommentResponse, error) {
	// todo: add your logic here and delete this line

	return &commentrpc.GetTopCommentResponse{}, nil
}
