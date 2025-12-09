// Package consts 提供 Errno 与错误信息映射
package consts

import "net/http"

const (
	ErrnoSuccess     = 0
	ErrnoUnknowError = 1

	// param error 1xxx
	ErrnoBindRequestError     = 1000
	ErrnoRequestValidateError = 1001

	// internal error 2xxx
	ErrnoInternalError = 2000
)

var ErrMsg = map[int]string{
	ErrnoSuccess:     "success",
	ErrnoUnknowError: "unknown error",

	ErrnoBindRequestError:     "bind request error",
	ErrnoRequestValidateError: "request validate error",

	ErrnoInternalError: "internal error",
}

// HTTPStatus 根据 Errno 返回对应的 HTTP 状态码
//
//   - 0 (ErrnoSuccess)     	→ 200
//   - 1 (ErrnoUnknowError) 	→ 500
//   - 1xxx (param error)   	→ 400
//   - 2xxx (internal error)	→ 500
//   - default     				→ 400
func HTTPStatus(errno int) int {
	switch {
	case errno == ErrnoSuccess:
		return http.StatusOK
	case errno == ErrnoUnknowError:
		return http.StatusInternalServerError
	case errno >= 1000 && errno < 2000:
		return http.StatusBadRequest
	case errno >= 2000 && errno < 3000:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}

}
