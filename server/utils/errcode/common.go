package errcode

import "net/http"

/*
通用的状态码
*/

const (
	ServerError = 500
)

// 业务基础状态码 code
const (
	Failed  = 1
	Success = 0
)

// 常用响应码 status
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
