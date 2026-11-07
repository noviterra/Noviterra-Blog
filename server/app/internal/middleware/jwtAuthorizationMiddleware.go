package middleware

import (
	"Noviterra/server/utils/jwtauth"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type identityRepository interface {
	GetUserAccountStatus(id string) (int, error)
	GetUserRole(id string) (int, error)
	CheckIDExists(id string) (bool, error)
}

type authMiddlewareCtxData interface {
	StoreUserIDInContext(ctx context.Context, userID string) context.Context
	StoreUserRoleInContext(ctx context.Context, userRole int) context.Context
}

type AuthMiddleware struct {
	authorization *jwtauth.Authentication
	identityRepo  identityRepository
	ctxdata       authMiddlewareCtxData
}

func NewAuthMiddleware(authorization *jwtauth.Authentication,
	identityRepo identityRepository, ctxdata authMiddlewareCtxData) *AuthMiddleware {
	return &AuthMiddleware{
		authorization: authorization,
		identityRepo:  identityRepo,
		ctxdata:       ctxdata,
	}
}

func (am *AuthMiddleware) JwtAuthMiddleware() rest.Middleware {
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
			claim, err := am.authorization.ParseAccessToken(tokenStr)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, fmt.Errorf("token解析失败：%v", err))
				return
			}
			if claim.ID == "" {
				httpx.ErrorCtx(r.Context(), w, errors.New("token中无有效信息！"))
				return
			}

			// 用户信息获取与校验
			idExist, err := am.identityRepo.CheckIDExists(claim.ID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, errors.New("用户校验失败！"))
				return
			}
			if !idExist {
				httpx.ErrorCtx(r.Context(), w, errors.New("不存在的账号！"))
				return
			}
			status, err := am.identityRepo.GetUserAccountStatus(claim.ID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, errors.New("状态校验失败！"))
				return
			}
			if status != 0 {
				httpx.ErrorCtx(r.Context(), w, errors.New("账号被封禁，操作失败！"))
				return
			}
			role, err := am.identityRepo.GetUserRole(claim.ID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, errors.New("身份校验失败！"))
				return
			}

			// 写入context中
			ctx := r.Context()
			ctx = am.ctxdata.StoreUserIDInContext(ctx, claim.ID)
			ctx = am.ctxdata.StoreUserRoleInContext(ctx, role)
			next(w, r.WithContext(ctx))
		}
	}
}
