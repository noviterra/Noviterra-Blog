// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package handler

import (
	"net/http"

	"Noviterra/server/app/agent/agentapi/internal/logic"
	"Noviterra/server/app/agent/agentapi/internal/svc"
	"Noviterra/server/app/agent/agentapi/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 配置站点助手
func DeployAgentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeployAgentRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewDeployAgentLogic(r.Context(), svcCtx)
		resp, err := l.DeployAgent(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
