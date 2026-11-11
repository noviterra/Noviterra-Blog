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
格式：3000--
*/
const (
	PasswordInvalidFormat = 30001
)
