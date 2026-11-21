// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"context"

	"Noviterra/server/app/agent/agentapi/internal/svc"
	"Noviterra/server/app/agent/agentapi/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeployAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 配置站点助手
func NewDeployAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeployAgentLogic {
	return &DeployAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeployAgentLogic) DeployAgent(req *types.DeployAgentRequest) (resp *types.Response, err error) {
	resp = new(types.Response)

	return
}
