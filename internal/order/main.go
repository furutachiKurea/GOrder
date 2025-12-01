package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/furutachiKurea/gorder/common/broker"
	"github.com/furutachiKurea/gorder/common/config"
	"github.com/furutachiKurea/gorder/common/discovery"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/common/logging"
	"github.com/furutachiKurea/gorder/common/server"
	"github.com/furutachiKurea/gorder/order/infrastructure/consumer"
	"github.com/furutachiKurea/gorder/order/ports"
	"github.com/furutachiKurea/gorder/order/service"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	logging.Init()
	if err := config.NewViperConfig(); err != nil {
		log.Fatal().Err(err).Msg("failed to init config")
	}

}

func main() {
	serviceName := viper.GetString("order.service-name")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, cleanup := service.NewApplication(ctx)
	defer cleanup()

	deregisterFn, err := discovery.RegisterToConsul(ctx, serviceName)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to register service %s to consul", serviceName)
	}
	defer func() { _ = deregisterFn() }()

	ch, closeCoon := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)
	defer func() {
		_ = ch.Close()
		_ = closeCoon()
	}()

	go consumer.NewConsumer(app).Listen(ch)

	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		svc := ports.NewGRPCServer(app)
		orderpb.RegisterOrderServiceServer(server, svc)
	})

	go server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		router.StaticFile("/success", "../../public/success.html")
		ports.RegisterHandlersWithOptions(router, HTTPServer{
			app: app,
		}, ports.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	<-ctx.Done()
}
