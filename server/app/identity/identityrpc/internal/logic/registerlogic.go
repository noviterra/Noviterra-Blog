package logic

import (
	"context"

	"Noviterra/server/app/identity/identityrpc/internal/svc"
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"

	"go.uber.org/zap"
)

type (
	registerIdentityCache interface {
		GetVerificationCode(ctx context.Context, email, t string) (string, error)
		DeleteVerificationCode(ctx context.Context, email, t string) error
	}

	registerIndetityRepository interface {
		CheckIDExists(ctx context.Context, id string) (bool, error)
		CheckEmailExists(ctx context.Context, email string) (bool, error)
		InsertUsersAccount(ctx context.Context,
			userAccount model.UserAccount, userInfo model.UserInfo) error
	}

	registerIdentityValidate interface {
		ValidatePasswordFormat(password string) (string, error)
	}

	registerCrypt interface {
		HashPassword(password string) (string, error)
	}

	registerIdentityResponse interface {
		SuccessResponse(resp *identityrpc.Response, msg string) (*identityrpc.Response, error)
		FailResponse(resp *identityrpc.Response, msg string, status int32, code int64) (*identityrpc.Response, error)
		FailResponseWithErr(resp *identityrpc.Response, err error) (*identityrpc.Response, error)
	}

	registerLogger interface {
		Info(msg string, fields ...zap.Field)
		Error(msg string, fields ...zap.Field)
	}
)

type RegisterLogic struct {
	ctx                context.Context
	noticeCache        registerIdentityCache
	identityRepository registerIndetityRepository
	identityValidate   registerIdentityValidate
	crypt              registerCrypt
	identityReponse    registerIdentityResponse
	logger             registerLogger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:                ctx,
		noticeCache:        svcCtx.NoticeCache,
		identityRepository: svcCtx.IdentityRepository,
		identityValidate:   svcCtx.IdentityValidate,
		crypt:              svcCtx.Crypt,
		identityReponse:    svcCtx.IdentityResponse,
		logger:             svcCtx.Logger,
	}
}

// 注册账号
func (l *RegisterLogic) Register(in *identityrpc.RegisterRequest) (*identityrpc.Response, error) {
	// todo: add your logic here and delete this line
	// fmt.Println("======server/app/user/rpc/registerlogic.go======")
	resp := new(identityrpc.Response)
	// 参数校验 账号唯一性 邮箱是否使用 验证码是否正确 密码格式限制（大小写数字与字符） 两次输入密码是否一致
	// 验证码获取与校验
	verificationCode, err := l.noticeCache.GetVerificationCode(l.ctx, in.Email, "register")
	if err != nil {
		l.logger.Error("缓存操作失败：查询注册验证码失败",
			zap.String("email", in.Email),
			zap.Error(err),
		)
		return l.identityReponse.FailResponse(resp, "缓存操作异常！",
			errcode.InternalError, errcode.GetAuthVerificationCodeFail)
	}
	if verificationCode == "" {
		l.logger.Error("读取到的注册验证码为空",
			zap.String("email", in.Email),
		)
		return l.identityReponse.FailResponse(resp, "无效的验证码，请重新发送！",
			errcode.InternalError, errcode.VerificationCodeExpired)
	}
	if in.VerificationCode != verificationCode {
		l.logger.Error("输入验证码与缓存验证码不同",
			zap.String("email", in.Email),
		)
		return l.identityReponse.FailResponse(resp, "验证码有误，请检查后重新输入！",
			errcode.InternalError, errcode.VerificationCodeInvalid)
	}

	// 账号查重
	idExist, err := l.identityRepository.CheckIDExists(l.ctx, in.ID)
	if err != nil {
		l.logger.Error("数据库操作失败：校验用户账号是否存在失败",
			zap.String("user-info", in.ID),
			zap.Error(err),
		)
		return l.identityReponse.FailResponse(resp, "数据库操作异常，请联系管理员！",
			errcode.InternalError, errcode.GetAuthInfoFail)
	}
	if idExist {
		l.logger.Error("用户账号已存在",
			zap.String("user-info", in.ID),
		)
		return l.identityReponse.FailResponse(resp, "账号重复，请检查后查询输入！",
			errcode.InternalError, errcode.AccountExists)
	}

	// 邮箱查重
	emailExist, err := l.identityRepository.CheckEmailExists(l.ctx, in.Email)
	if err != nil {
		l.logger.Error("数据库操作失败：校验用户邮箱是否存在失败",
			zap.String("email", in.Email),
			zap.Error(err),
		)
		return l.identityReponse.FailResponse(resp, "数据库操作异常，请联系管理员！",
			errcode.InternalError, errcode.GetAuthInfoFail)
	}
	if emailExist {
		l.logger.Error("用户邮箱已被其他账号绑定",
			zap.String("email", in.Email),
		)
		return l.identityReponse.FailResponse(resp, "该邮箱已绑定其他账号，请检查后重新注册！",
			errcode.InternalError, errcode.EmailExists)
	}

	// 密码校验
	if in.Password != in.VerifPassword {
		l.logger.Error("两次输入的密码不一致",
			zap.String("user-info", in.ID),
		)
		return l.identityReponse.FailResponse(resp, "两次密码输入不一致，请检查后重试！",
			errcode.InternalError, errcode.PasswordNotMatch)
	}
	_, err = l.identityValidate.ValidatePasswordFormat(in.Password)
	if err != nil {
		l.logger.Error("校验密码格式失败",
			zap.String("user-info", in.ID),
			zap.Error(err),
		)
		return l.identityReponse.FailResponse(resp, "密码格式错误，请检查后重试！",
			errcode.InternalError, errcode.PasswordInvalidFormat)
	}

	// 密码加密
	hashPwd, err := l.crypt.HashPassword(in.Password)
	if err != nil {
		l.logger.Error("密码加密失败",
			zap.String("user-info", in.ID),
			zap.Error(err),
		)
		return l.identityReponse.FailResponse(resp, "数据加密异常，请联系管理员！",
			errcode.InternalError, errcode.PasswordHashFail)
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
	err = l.identityRepository.InsertUsersAccount(l.ctx, userAccount, userInfo)
	if err != nil {
		l.logger.Error("数据库操作失败：用户账号存储失败",
			zap.String("user-info", in.ID),
			zap.String("email", in.Email),
			zap.Error(err),
		)
		return l.identityReponse.FailResponse(resp, "数据库操作异常！",
			errcode.InternalError, errcode.RegisterFail)
	}

	// 注册成功后删除缓存中的验证码
	err = l.noticeCache.DeleteVerificationCode(l.ctx, in.Email, "register")
	if err != nil { // 删除失败不返回错误 后续会有特殊处理
		l.logger.Error("缓存操作失败：删除注册验证码失败",
			zap.String("email", in.Email),
			zap.Error(err),
		)
	}

	l.logger.Info("用户账号注册成功",
		zap.String("user-info", in.ID),
		zap.String("email", in.Email),
	)
	return l.identityReponse.SuccessResponse(resp, "账号注册成功！")
}
