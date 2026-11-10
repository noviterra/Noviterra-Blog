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
格式：2000--
*/
const (
	BadPageData = 200001
)

/*
认证模块服务状态码
格式：3000--
*/
const (
	PasswordInvalidFormat = 30001
)
