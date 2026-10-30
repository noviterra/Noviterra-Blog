package interceptor

import (
	"Noviterra/server/utils/ctxdata"
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServiceTokenConfig struct {
	SecretToken string
	SkipToken   string
}
type ServiceTokenInterceptor struct {
	ctxData *ctxdata.CtxData
	conf    ServiceTokenConfig
}

func NewServiceTokenInterceptor(ctxdata *ctxdata.CtxData, serviceTokenConfig ServiceTokenConfig) *ServiceTokenInterceptor {
	return &ServiceTokenInterceptor{
		ctxData: ctxdata,
		conf:    serviceTokenConfig,
	}
}

// 白名单 不做校验
var skipMethods = map[string]bool{
	"/grpc.health.v1.Health/Check": true,
	"/grpc.health.v1.Health/Watch": true,
}

func (si *ServiceTokenInterceptor) ServiceTokenInterceptor(secretToken, skipToken string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// 白名单跳过
		if skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// 将metadata中的信息取出
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("metadata数据丢失！")
		}

		// 校验服务令牌
		secretTokens := md.Get("service-token")
		if len(secretTokens) <= 0 || secretTokens[0] != secretToken {
			return nil, errors.New("无效的service-token！")
		}

		// 校验是否提取ID的令牌：SkipToken
		skipTokens := md.Get("skip-token")
		if len(skipTokens) <= 0 || skipTokens[0] != skipToken { // 无跳过令牌 继续提取ID
			// 取出user-id
			ids := md.Get("user-id")
			if len(ids) <= 0 {
				return nil, errors.New("user-id丢失！")
			}
			id := ids[0]
			if len(id) <= 0 {
				return nil, errors.New("无效的user-id！")
			}
			ctx = si.ctxData.StoreUserIDInContext(ctx, id)
			return handler(ctx, req)
		}
		// 存在正确的跳过令牌 直接返回ctx 不写入ID
		return handler(ctx, req)
	}
}

func (si *ServiceTokenInterceptor) GetServiceTokenInConf() string {
	return si.conf.SecretToken
}

func (si *ServiceTokenInterceptor) GetSkipTokenInConf() string {
	return si.conf.SkipToken
}
