package middleware

import (
	"Noviterra/server/utils/jwtauth"
	"context"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest"
)

/*
检查用户的登录状态 不要求必须登录 但登录的话就进行身份校验
使用该中间件限制未登录用户的个别操作 如查看全部评论 查看完整文章等等...
*/

type hasAuthTokenCtxData interface {
	SetUserLoginStatus(ctx context.Context, isLogin bool) context.Context
	StoreUserIDInContext(ctx context.Context, userID string) context.Context
}

type hasAuthTokenAuthentication interface {
	ParseAccessToken(tokenStr string) (*jwtauth.UserClaims, error)
}

type hasAuthTokenIdentityRepository interface {
	GetUserAccountStatus(ctx context.Context, id string) (int, error)
	CheckIDExists(ctx context.Context, id string) (bool, error)
}

type HasAuthTokenMiddleware struct {
	ctxdata            hasAuthTokenCtxData
	authorization      hasAuthTokenAuthentication
	identityRepository hasAuthTokenIdentityRepository
}

func (hm *HasAuthTokenMiddleware) HasAuthToken() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = hm.ctxdata.SetUserLoginStatus(ctx, false) // 默认游客
			var userID string

			// 错误位置只需要return 使用defer替换掉大多数重复执行的语句
			defer func() {
				if userID != "" {
					ctx = hm.ctxdata.StoreUserIDInContext(ctx, userID)
					ctx = hm.ctxdata.SetUserLoginStatus(ctx, true)
				}
				next(w, r.WithContext(ctx))
			}()

			// 从请求头中获取Token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" { // 为空则为未登录 设置为游客状态
				return
			}

			// 校验Token 校验失败、Token过期与失效都降级为游客状态 不直接return 仅日志记录
			// 校验格式 Bearer <token>
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return
			}
			tokenStr := parts[1]

			// 解析Token--AccessToken
			claim, err := hm.authorization.ParseAccessToken(tokenStr)
			if err != nil {
				return
			}
			if claim.ID == "" {
				return
			}

			// 用户信息获取与校验
			idExist, err := hm.identityRepository.CheckIDExists(ctx, claim.ID)
			if err != nil {
				return
			}
			if !idExist {
				return
			}
			status, err := hm.identityRepository.GetUserAccountStatus(ctx, claim.ID)
			if err != nil {
				return
			}
			if status != 0 {
				return
			}

			// 全部通过 登录状态正确
			userID = claim.ID
		}
	}
}
