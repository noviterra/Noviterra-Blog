package logic

import (
	"context"
	"fmt"

	"Noviterra/server/app/model"
	"Noviterra/server/app/user/userrpc/internal/svc"
	"Noviterra/server/app/user/userrpc/types/userrpc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// rpc方法
func (l *RegisterLogic) Register(in *userrpc.RegisterRequest) (*userrpc.RegisterResponse, error) {
	// todo: add your logic here and delete this line
	// fmt.Println("======server/app/user/rpc/registerlogic.go======")

	// 参数校验 账号唯一性 邮箱是否使用 验证码是否正确 密码格式限制（大小写数字与字符） 两次输入密码是否一致
	// 验证码获取与校验
	verificationCode, err := l.svcCtx.RDB.GetVerificationCode(l.ctx, in.Email, "register")
	fmt.Printf("JSON Data:%s, Real Data:%s\n", in.VerificationCode, verificationCode)
	if err != nil {
		return &userrpc.RegisterResponse{
			Code:    1,
			Message: "Redis缓存异常，请联系管理员！",
		}, err
	}
	if verificationCode == "" {
		return &userrpc.RegisterResponse{
			Code:    1,
			Message: "无效的验证码，请重新发送！",
		}, nil
	}
	if in.VerificationCode != verificationCode {
		return &userrpc.RegisterResponse{
			Code:    1,
			Message: "验证码有误，请检查后重新输入！",
		}, nil
	}
	l.svcCtx.Logger.Info("验证码校验成功！")
	// 账号查重
	idExist, err := l.svcCtx.DB.CheckIDExists(in.ID)
	if err != nil {
		return &userrpc.RegisterResponse{
			Code:    1,
			Message: "数据库操作异常，请联系管理员！",
		}, err
	}
	if idExist {
		return &userrpc.RegisterResponse{
			Code:    1,
			Message: "账号重复，请检查后查询输入！",
		}, nil
	}
	l.svcCtx.Logger.Info("账号校验成功！")

	// 邮箱查重
	emailExist, err := l.svcCtx.DB.CheckEmailExists(in.Email)
	if err != nil {
		return &userrpc.RegisterResponse{
			Code:    1,
			Message: "数据库操作异常，请联系管理员！",
		}, err
	}
	if emailExist {
		return &userrpc.RegisterResponse{
			Code:    1,
			Message: "该邮箱已绑定其他账号，请检查后重新注册！",
		}, nil
	}
	l.svcCtx.Logger.Info("邮箱校验成功！")

	// 密码校验
	if in.Password != in.VerifPassword {
		return &userrpc.RegisterResponse{
			Code:    1,
			Message: "两次密码输入不一致，请检查后重试！",
		}, nil
	}
	passwordAvailable := l.svcCtx.CommonUtils.CheckPasswordFormat(in.Password)
	if !passwordAvailable {
		return &userrpc.RegisterResponse{
			Code:    1,
			Message: "密码格式错误，请检查后重试！",
		}, nil
	}
	l.svcCtx.Logger.Info("密码校验成功！")

	// 密码加密
	hashPwd, err := l.svcCtx.Crypt.HashPassword(in.Password)
	if err != nil {
		return &userrpc.RegisterResponse{
			Code:    1,
			Message: "数据加密异常，请联系管理员！",
		}, err
	}

	// 账号存储到数据库中 并且创建用户实例存入到数据库中
	userAccount := model.UserAccount{
		ID:       in.ID,
		Password: hashPwd,
		Email:    in.Email,
	}
	userInfo := model.UserInfo{
		ID:    in.ID,
		Name:  in.ID,
		Email: in.Email,
	}
	err = l.svcCtx.DB.InsertUsersAccount(userAccount, userInfo)
	if err != nil {
		return &userrpc.RegisterResponse{
			Code:    1,
			Message: "数据库写入失败，请联系管理员！",
		}, err
	}

	l.svcCtx.Logger.Info("账号注册成功！")

	// 注册成功后删除缓存中的验证码
	err = l.svcCtx.RDB.DeleteVerificationCode(l.ctx, in.Email, "register")
	if err != nil {
		l.svcCtx.Logger.Error("验证码缓存清理失败！")
	}
	l.svcCtx.Logger.Info("验证码缓存清理成功！")

	return &userrpc.RegisterResponse{
		Code:    0,
		Message: "账号注册成功！",
	}, nil
}
