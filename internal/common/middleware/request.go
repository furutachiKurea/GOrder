package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func RequestLog(l zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestIn(c, l)
		defer requestOut(c, l)
		c.Next()
	}
}

func requestIn(c *gin.Context, l zerolog.Logger) {
	c.Set("request_start", time.Now())

	body := c.Request.Body
	bodyBytes, _ := io.ReadAll(body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var compactJson bytes.Buffer
	_ = json.Compact(&compactJson, bodyBytes)

	l.Info().
		Str("uri", c.Request.RequestURI).
		Str("method", c.Request.Method).
		Str("from", c.RemoteIP()).
		Time("start", time.Now()).
		Str("body", compactJson.String()).
		Msg("__request_in")

}

func requestOut(c *gin.Context, l zerolog.Logger) {
	resp, _ := c.Get("response")
	start, _ := c.Get("request_start")
	startTime := start.(time.Time)

	// response 可能未设置（如 404 或中间件提前终止），避免对 nil 做断言
	respBytes, _ := resp.([]byte)
	respStr := ""
	if respBytes != nil {
		respStr = string(respBytes)
	}

	l.Info().
		Int("status_code", c.Writer.Status()).
		Int64("proc_time_ms", time.Since(startTime).Milliseconds()).
		Str("response", respStr).
		Msg("__request_out")
}
