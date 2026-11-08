package logic

import (
	"Noviterra/server/app/identity/identityrpc/internal/svc"
	"Noviterra/server/app/identity/identityrpc/types/identityrpc"
	"Noviterra/server/app/model"
	"Noviterra/server/utils/errcode"
	"context"
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
		CheckPasswordFormat(password string) bool
	}

	registerCrypt interface {
		HashPassword(password string) (string, error)
	}

	registerIdentityResponse interface {
		SuccessResponse(resp *identityrpc.Response, msg string) (*identityrpc.Response, error)
		FailResponse(resp *identityrpc.Response, msg string, status int32, code int64) (*identityrpc.Response, error)
		FailResponseWithErr(resp *identityrpc.Response, err error) (*identityrpc.Response, error)
	}
)

type RegisterLogic struct {
	ctx                context.Context
	noticeCache        registerIdentityCache
	identityRepository registerIndetityRepository
	identityValidate   registerIdentityValidate
	crypt              registerCrypt
	identityReponse    registerIdentityResponse
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:                ctx,
		noticeCache:        svcCtx.NoticeCache,
		identityRepository: svcCtx.IdentityRepository,
		identityValidate:   svcCtx.IdentityValidate,
		crypt:              svcCtx.Crypt,
		identityReponse:    svcCtx.IdentityResponse,
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
		return l.identityReponse.FailResponse(resp, "缓存操作异常！", errcode.InternalError, errcode.Failed)
	}
	if verificationCode == "" {
		return l.identityReponse.FailResponse(resp, "无效的验证码，请重新发送！", errcode.InternalError, errcode.Failed)
	}
	if in.VerificationCode != verificationCode {
		return l.identityReponse.FailResponse(resp, "验证码有误，请检查后重新输入！", errcode.InternalError, errcode.Failed)

	}
	// 账号查重
	idExist, err := l.identityRepository.CheckIDExists(l.ctx, in.ID)
	if err != nil {
		return l.identityReponse.FailResponse(resp, "数据库操作异常，请联系管理员！", errcode.InternalError, errcode.Failed)

	}
	if idExist {
		return l.identityReponse.FailResponse(resp, "账号重复，请检查后查询输入！", errcode.InternalError, errcode.Failed)

	}

	// 邮箱查重
	emailExist, err := l.identityRepository.CheckEmailExists(l.ctx, in.Email)
	if err != nil {
		return l.identityReponse.FailResponse(resp, "数据库操作异常，请联系管理员！", errcode.InternalError, errcode.Failed)
	}
	if emailExist {
		return l.identityReponse.FailResponse(resp, "该邮箱已绑定其他账号，请检查后重新注册！", errcode.InternalError, errcode.Failed)
	}

	// 密码校验
	if in.Password != in.VerifPassword {
		return l.identityReponse.FailResponse(resp, "两次密码输入不一致，请检查后重试！", errcode.InternalError, errcode.Failed)
	}
	passwordAvailable := l.identityValidate.CheckPasswordFormat(in.Password)
	if !passwordAvailable {
		return l.identityReponse.FailResponse(resp, "密码格式错误，请检查后重试！", errcode.InternalError, errcode.Failed)
	}

	// 密码加密
	hashPwd, err := l.crypt.HashPassword(in.Password)
	if err != nil {
		return l.identityReponse.FailResponse(resp, "数据加密异常，请联系管理员！", errcode.InternalError, errcode.Failed)
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
		return l.identityReponse.FailResponse(resp, "数据库操作异常！", errcode.InternalError, errcode.Failed)
	}

	// 注册成功后删除缓存中的验证码
	err = l.noticeCache.DeleteVerificationCode(l.ctx, in.Email, "register")
	if err != nil { // 删除失败不返回错误 后续会有特殊处理
	}

	return l.identityReponse.SuccessResponse(resp, "账号注册成功！")
}
