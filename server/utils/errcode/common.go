package errcode

import "net/http"

// RequestSuccess
// ====================================== Status ================================================
// 请求完成
const (
	RequestSuccess = http.StatusOK // 200
)

// 客户端错误
const (
	InvalidParam   = http.StatusBadRequest          // 400
	Unauthorized   = http.StatusUnauthorized        // 401
	Forbidden      = http.StatusForbidden           // 403
	NotFound       = http.StatusNotFound            // 404
	MethodNotAllow = http.StatusMethodNotAllowed    // 405
	Conflict       = http.StatusConflict            // 409
	Unprocessable  = http.StatusUnprocessableEntity // 422
	TooManyRequest = http.StatusTooManyRequests     // 429
)

// 服务端错误
const (
	InternalError  = http.StatusInternalServerError // 500
	BadGateway     = http.StatusBadGateway          // 502
	ServiceUnavail = http.StatusServiceUnavailable  // 503
	GatewayTimeout = http.StatusGatewayTimeout      // 504
)

// ====================================== Code ================================================
// 业务基础状态码 code
const (
	Failed  = 1
	Success = 0
)

/*
用户通用状态码
格式：100--
*/
const (
	GetUserInfoFail              = 1001
	ValidateUserInfoFail         = 1002
	GetVerificationCodeFail      = 1003
	ValidateVerificationCodeFail = 1004
	ValidateUserRequestParamFail = 1005
	InterCallToolFail            = 1006
	UserOperationOutOfBounds     = 1007
	GetContextUserIDFail         = 1008
	UserNotFound                 = 1009
	UserRoleInsufficient         = 1010
	UserSelfOperationForbidden   = 1011
	GetUserRoleFail              = 1012
	UpdateUserInfoFail           = 1013
	UserUpdatePasswordFail       = 1014
	HashPasswordFail             = 1015
	CheckPasswordFail            = 1016
	UserPasswordIncorrect        = 1017
	BlackUserFail                = 1018
	UnBlackUserFail              = 1019
	UserInfoEmpty                = 1020
	UserEmailMismatch            = 1021
	VerificationCodeNotFound     = 1022
	GetUserCacheFail             = 1023
	GetUserListFail              = 1024
)

/*
文章通用状态码
格式：200--
*/
const (
	GetArticleInfoFail              = 2001
	GetArticleListFail              = 2002
	ValidateArticleIDFail           = 2003
	ValidateArticleRequestParamFail = 2004
	ValidateArticleCoverFail        = 2005
	GetArticleArchiveFail           = 2006
	GetArticleCacheFail             = 2007
	PublishArticleFail              = 2008
	UpdateArticleFail               = 2009
	DeleteArticleFail               = 2010
	SearchArticleFail               = 2011
	UpdateArticleCoverFail          = 2012
	ArticleNotFound                 = 2013
	ArticleOperationOutOfBounds     = 2014
)

/*
认证模块服务状态码
格式：300--
*/
const (
	PasswordInvalidFormat        = 3001
	PasswordNotMatch             = 3002
	AccountNotFound              = 3003
	AccountExists                = 3004
	EmailNotFound                = 3005
	EmailExists                  = 3006
	EmailMismatch                = 3007
	VerificationCodeExpired      = 3008
	VerificationCodeInvalid      = 3009
	GetAuthVerificationCodeFail  = 3010
	GetAuthInfoFail              = 3011
	GenerateTokenFail            = 3012
	SaveTokenFail                = 3013
	RefreshTokenInvalid          = 3014
	PasswordIncorrect            = 3015
	PasswordHashFail             = 3016
	PasswordCheckFail            = 3017
	RegisterFail                 = 3018
	UpdatePasswordFail           = 3019
	IdentityOperationOutOfBounds = 3020
)

/*
互动模块服务状态码
格式：400--
*/
const (
	CommentUserInfoInvalid     = 4001
	CommentActionInvalid       = 4002
	CollectArticleFail         = 4003
	ArticleAlreadyCollected    = 4004
	UncollectArticleFail       = 4005
	ArticleNotCollected        = 4006
	CreateCommentFail          = 4007
	GetTopCommentCacheFail     = 4008
	GetTopCommentFail          = 4009
	GetTopCommentUserInfoFail  = 4010
	GetTreeCommentCacheFail    = 4011
	GetTreeCommentFail         = 4012
	GetTreeCommentUserInfoFail = 4013
	LikeArticleFail            = 4014
	ArticleAlreadyLiked        = 4015
	UnlikeArticleFail          = 4016
	ArticleNotLiked            = 4017
)
