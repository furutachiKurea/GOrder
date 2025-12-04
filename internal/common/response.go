package common

import (
	"encoding/json"
	"net/http"

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
	r := response{
		Errno:   0,
		Message: "success",
		Data:    data,
		TraceID: tracing.TraceID(c.Request.Context()),
	}

	resp, _ := json.Marshal(r)
	c.Set("response", resp)
	c.JSON(http.StatusOK, r)
}

func (b *BaseResponse) error(c *gin.Context, err error) {
	r := response{
		Errno:   2,
		Message: err.Error(),
		Data:    nil,
		TraceID: tracing.TraceID(c.Request.Context()),
	}

	resp, _ := json.Marshal(r)
	c.Set("response", resp)
	c.JSON(http.StatusOK, r)
}
