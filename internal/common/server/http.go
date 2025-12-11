package server

import (
	"github.com/furutachiKurea/gorder/common/middleware"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func RunHTTPServer(serviceName string, wrapper func(router *gin.Engine)) {
	addr := viper.Sub(serviceName).Get("http-addr")
	if addr == nil {
		addr = viper.Get("fallback-http-addr")
		log.Warn().Str("fallback", addr.(string)).Msg("http server addr is nil, use fallback-http-addr instead")
	}

	RunHTTPServerOnAddr(addr.(string), wrapper)
}

func RunHTTPServerOnAddr(addr string, wrapper func(router *gin.Engine)) {
	apiRouter := gin.New()
	setMiddlewares(apiRouter)
	wrapper(apiRouter)
	if err := apiRouter.Run(addr); err != nil {
		panic(err)
	}
}

func setMiddlewares(r *gin.Engine) {
	r.Use(middleware.RequestLog(log.Logger))
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("default_server"))
	r.Use()
}
