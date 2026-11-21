package logic

import (
	"context"

	"Noviterra/server/app/agent/agentrpc/internal/svc"
	"Noviterra/server/app/agent/agentrpc/types/agentrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeployLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeployLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeployLogic {
	return &DeployLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 配置对话助手
func (l *DeployLogic) Deploy(in *agentrpc.DeployAgentRequest) (*agentrpc.Response, error) {
	// todo: add your logic here and delete this line

	return &agentrpc.Response{}, nil
}
