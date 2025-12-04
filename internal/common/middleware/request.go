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
		Time("start", time.Now()).
		Str("body", compactJson.String()).
		Str("from", c.RemoteIP()).
		Str("uri", c.Request.RequestURI).
		Msg("__request_in")

}

func requestOut(c *gin.Context, l zerolog.Logger) {
	resp, _ := c.Get("response")
	start, _ := c.Get("request_start")
	startTime := start.(time.Time)

	l.Info().
		Int("status_code", c.Writer.Status()).
		Int64("proc_time_ms", time.Since(startTime).Milliseconds()).
		Str("response", string(resp.([]byte))).
		Msg("__request_out")
}
