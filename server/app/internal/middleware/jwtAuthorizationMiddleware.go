package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type ctxKeyID string
type ctxKeyRole string

const (
	ctxKeyUserID   ctxKeyID   = "userId"
	ctxKeyUserRole ctxKeyRole = "userRole"
)

func (m *Middleware) JwtAuthMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 从请求头中获取Token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httpx.ErrorCtx(r.Context(), w, errors.New("无请求头！"))
				return
			}

			// 校验格式 Bearer <token>
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				httpx.ErrorCtx(r.Context(), w, errors.New("请求头格式错误！"))
				return
			}
			tokenStr := parts[1]

			// 解析Token--AccessToken
			claim, err := m.authorization.ParseAccessToken(tokenStr)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, fmt.Errorf("Token解析失败：%v", err))
				return
			}
			if claim.ID == "" {
				httpx.ErrorCtx(r.Context(), w, errors.New("Token中无有效信息！"))
				return
			}

			// todo: 用户状态校验 账号校验与身份校验
			idExist, err := m.userRepo.CheckIDExists(claim.ID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, errors.New("用户校验失败！"))
				return
			}
			if !idExist {
				httpx.ErrorCtx(r.Context(), w, errors.New("疑似伪造用户，劝你心善！"))
				return
			}
			status, err := m.userRepo.GetUserAccountStatus(claim.ID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, errors.New("状态校验失败！"))
				return
			}
			if status != 0 {
				httpx.ErrorCtx(r.Context(), w, errors.New("账号被封禁，操作失败！"))
				return
			}
			role, err := m.userRepo.GetUserRole(claim.ID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, errors.New("身份校验失败！"))
				return
			}

			// 写入context中
			ctx := r.Context()
			ctx = context.WithValue(r.Context(), ctxKeyUserID, claim.ID)
			ctx = context.WithValue(ctx, ctxKeyUserRole, int(role))
			next(w, r.WithContext(ctx))
		}
	}
}

// 从context中获取用户ID
func (m *Middleware) GetUserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyUserID).(string); ok {
		return v
	}
	return ""
}

// 从context中获取用户身份
func (m *Middleware) GetUserRoleFromContext(ctx context.Context) int {
	if v, ok := ctx.Value(ctxKeyUserRole).(int); ok {
		return v
	}
	return -1
}
