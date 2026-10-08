package logic

import (
	"context"
	"fmt"

	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendVerificationCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendVerificationCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendVerificationCodeLogic {
	return &SendVerificationCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 邮箱辅助功能
func (l *SendVerificationCodeLogic) SendVerificationCode(in *userrpc.SendVerificationCodeRequest) (*userrpc.SendVerificationCodeResponse, error) {
	// todo: add your logic here and delete this line
	// fmt.Println("======server/app/user/rpc/SendVerificationCodelogic.go======")

	// 检查邮箱是否存在
	emailExist, err := l.svcCtx.DB.CheckEmailExists(in.Email)
	if err != nil {
		return &userrpc.SendVerificationCodeResponse{
			Code:    1,
			Message: "数据库异常！",
		}, err
	}
	// 对于注册 需要保证邮箱不存在 即没有注册过账号
	if in.Type == "register" {
		if emailExist {
			return &userrpc.SendVerificationCodeResponse{
				Code:    1,
				Message: "邮箱已经注册过账号！",
			}, nil
		}
	} else {
		// 对于其他则需要保证邮箱已经注册过账号
		if !emailExist {
			return &userrpc.SendVerificationCodeResponse{
				Code:    1,
				Message: "邮箱尚未注册账号！",
			}, nil
		}
	}

	// 缓存查询 限制一分钟内只可以发送一次
	ok, err := l.svcCtx.RDB.LimitVerificationCode(l.ctx, in.Email)
	if err != nil {
		return &userrpc.SendVerificationCodeResponse{
			Code:    1,
			Message: "验证码发送失败!",
		}, err
	}
	if !ok {
		fmt.Println(1)
		return &userrpc.SendVerificationCodeResponse{
			Code:    1,
			Message: "验证码发送频繁,请稍后再试!",
		}, nil
	}

	// 生成验证码
	verificationCode, err := l.svcCtx.CommonUtils.GenerateVerificationCode()
	if err != nil {
		return &userrpc.SendVerificationCodeResponse{
			Code:    1,
			Message: "验证码生成失败!",
		}, err
	}
	// 存入缓存中 五分钟有效
	err = l.svcCtx.RDB.StoreVerificationCode(l.ctx, in.Email, in.Type, verificationCode)
	if err != nil {
		return &userrpc.SendVerificationCodeResponse{
			Code:    1,
			Message: "缓存存储失败，请联系管理员!",
		}, err
	}

	// 发送验证码
	emailSubject, emailBody := l.svcCtx.EmailSender.TextEmailEditor(verificationCode, in.Type)
	if emailSubject == "" || emailBody == "" {
		return &userrpc.SendVerificationCodeResponse{
			Code:    1,
			Message: "异常请求！",
		}, nil
	}
	err = l.svcCtx.EmailSender.SendTextEmails([]string{in.Email}, emailSubject, emailBody)
	if err != nil {
		return &userrpc.SendVerificationCodeResponse{
			Code:    1,
			Message: "验证码发送失败!",
		}, err
	}

	return &userrpc.SendVerificationCodeResponse{
		Code:    0,
		Message: "验证码发送成功!",
	}, nil
}
