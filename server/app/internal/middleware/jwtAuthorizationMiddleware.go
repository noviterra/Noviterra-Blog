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
	GetUserAccountStatus(ctx context.Context, id string) (int, error)
	GetUserRole(ctx context.Context, id string) (int, error)
	CheckIDExists(ctx context.Context, id string) (bool, error)
}

type authMiddlewareCtxData interface {
	StoreUserIDInContext(ctx context.Context, userID string) context.Context
	StoreUserRoleInContext(ctx context.Context, userRole int) context.Context
}

type authMiddlewareAuthorization interface {
	ParseAccessToken(tokenStr string) (*jwtauth.UserClaims, error)
}

type AuthMiddleware struct {
	authorization authMiddlewareAuthorization
	identityRepo  identityRepository
	ctxdata       authMiddlewareCtxData
}

func NewAuthMiddleware(authorization authMiddlewareAuthorization,
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
			ctx := r.Context()
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
			idExist, err := am.identityRepo.CheckIDExists(ctx, claim.ID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, errors.New("用户校验失败！"))
				return
			}
			if !idExist {
				httpx.ErrorCtx(r.Context(), w, errors.New("不存在的账号！"))
				return
			}
			status, err := am.identityRepo.GetUserAccountStatus(ctx, claim.ID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, errors.New("状态校验失败！"))
				return
			}
			if status != 0 {
				httpx.ErrorCtx(r.Context(), w, errors.New("账号被封禁，操作失败！"))
				return
			}
			role, err := am.identityRepo.GetUserRole(ctx, claim.ID)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, errors.New("身份校验失败！"))
				return
			}

			// 写入context中
			ctx = am.ctxdata.StoreUserIDInContext(ctx, claim.ID)
			ctx = am.ctxdata.StoreUserRoleInContext(ctx, role)
			next(w, r.WithContext(ctx))
		}
	}
}
