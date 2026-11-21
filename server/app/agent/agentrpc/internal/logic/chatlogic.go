package logic

import (
	"context"

	"Noviterra/server/app/agent/agentrpc/internal/svc"
	"Noviterra/server/app/agent/agentrpc/types/agentrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 对话
func (l *ChatLogic) Chat(in *agentrpc.ChatRequest) (*agentrpc.ChatResponse, error) {
	// todo: add your logic here and delete this line

	return &agentrpc.ChatResponse{}, nil
}
