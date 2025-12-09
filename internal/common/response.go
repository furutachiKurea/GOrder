package common

import (
	"encoding/json"

	"github.com/furutachiKurea/gorder/common/consts"
	"github.com/furutachiKurea/gorder/common/handler/errors"
	"github.com/furutachiKurea/gorder/common/tracing"

	"github.com/gin-gonic/gin"
)

type response struct {
	Errno   int    `json:"errno"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
	Data    any    `json:"data"`
}

type BaseResponse struct{}

func (b *BaseResponse) Response(c *gin.Context, err error, data any) {
	if err != nil {
		b.error(c, err)
	} else {
		b.success(c, data)
	}
}

func (b *BaseResponse) success(c *gin.Context, data any) {
	errno, errMsg := errors.Output(nil)
	r := response{
		Errno:   errno,
		Message: errMsg,
		Data:    data,
		TraceID: tracing.TraceID(c.Request.Context()),
	}

	resp, _ := json.Marshal(r)
	c.Set("response", resp)
	c.JSON(consts.HTTPStatus(errno), r)
}

func (b *BaseResponse) error(c *gin.Context, err error) {
	errno, errMsg := errors.Output(err)
	r := response{
		Errno:   errno,
		Message: errMsg,
		Data:    nil,
		TraceID: tracing.TraceID(c.Request.Context()),
	}

	resp, _ := json.Marshal(r)
	c.Set("response", resp)
	c.JSON(consts.HTTPStatus(errno), r)
}
