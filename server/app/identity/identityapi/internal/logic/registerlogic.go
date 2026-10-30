// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package logic

import (
	"Noviterra/server/app/identity/identityapi/internal/svc"
	"Noviterra/server/app/identity/identityapi/internal/types"
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/utils/errcode"
	"context"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type registerIdentityModel interface {
	Register(ctx context.Context, in *identityrpc.RegisterRequest, opts ...grpc.CallOption) (*identityrpc.Response, error)
}

type registerServiceTokenInterceptor interface {
	GetServiceTokenInConf() string
	GetSkipTokenInConf() string
}

type RegisterLogic struct {
	ctx                     context.Context
	identityModel           registerIdentityModel
	serviceTokenInterceptor registerServiceTokenInterceptor
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:                     ctx,
		identityModel:           svcCtx.IdentityModel,
		serviceTokenInterceptor: svcCtx.ServiceTokenInterceptor,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.Response, err error) {
	// todo: add your logic here and delete this line
	//fmt.Println("======server/app/user/api/registerlogic.go======")
	resp = new(types.Response)

	// binding tag验证
	if err := Validate.Struct(*req); err != nil {
		// 解析错误类型
		if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
			return failResponse(resp, TypeName[errs[0].Field()]+"格式错误！", errcode.InvalidParam, errcode.Failed)
		}
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}

	ctx := metadata.AppendToOutgoingContext(l.ctx,
		"service-token", l.serviceTokenInterceptor.GetServiceTokenInConf(),
		"skip-token", l.serviceTokenInterceptor.GetSkipTokenInConf(),
	)

	rep, err := l.identityModel.Register(ctx, &identityrpc.RegisterRequest{
		ID:               req.ID,
		Password:         req.Password,
		Email:            req.Email,
		VerifPassword:    req.VerifPassword,
		VerificationCode: req.VerificationCode,
	})

	if err != nil {
		return failResponseWithErr(resp, err, errcode.InvalidParam, errcode.Failed)
	}
	if rep.Code != 0 {
		return failResponse(resp, rep.Message, int(rep.Status), int(rep.Code))
	}

	return successResponse(resp, "账号注册成功！")
}
