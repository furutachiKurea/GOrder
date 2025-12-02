package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func StructuredLog(l zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		elapsed := time.Since(t)
		l.Info().
			Int64("time_elapsed_ms", elapsed.Milliseconds()).
			Str("request_uri", c.Request.RequestURI).
			Str("client_ip", c.ClientIP()).
			Str("full_path", c.FullPath()).
			Msg("request_out")

	}
}
