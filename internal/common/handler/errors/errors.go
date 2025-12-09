// Package errors 提供对外的统一错误封装
package errors

import (
	"errors"

	"github.com/furutachiKurea/gorder/common/consts"

	"github.com/rs/zerolog/log"
)

type Error struct {
	code int
	msg  string
	err  error
}

func NewWithError(code int, err error) *Error {
	return &Error{
		code: code,
		err:  err,
	}
}

func (e Error) Error() string {
	var msg string
	if e.msg != "" {
		msg = e.msg
	} else {
		msg = consts.ErrMsg[e.code]
	}

	if e.err == nil {
		return msg
	}

	return msg + ": " + e.err.Error()
}

// Errno 从 err 中提取 errno，如果无法提取则返回 consts.ErrnoUnknowError
func Errno(err error) int {
	if err == nil {
		return consts.ErrnoSuccess
	}

	target := &Error{}
	if errors.As(err, target) {
		log.Debug().Err(err).Msg("is errors.Error")
		return target.code
	}

	return consts.ErrnoUnknowError
}

// Output 解包 err, 返回 errno 和 msg
func Output(err error) (errno int, msg string) {
	if err == nil {
		return consts.ErrnoSuccess, consts.ErrMsg[consts.ErrnoSuccess]
	}

	errno = Errno(err)
	if errno == -1 {
		return consts.ErrnoUnknowError, consts.ErrMsg[consts.ErrnoUnknowError]
	}
	return errno, err.Error()
}
