package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/furutachiKurea/gorder/common/config"
	"github.com/furutachiKurea/gorder/common/discovery"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/common/server"
	"github.com/furutachiKurea/gorder/order/ports"
	"github.com/furutachiKurea/gorder/order/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	if err := config.NewViperConfig(); err != nil {
		logrus.Fatal(err)
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
		logrus.Fatal(err)
	}
	defer func() { _ = deregisterFn() }()

	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		svc := ports.NewGRPCServer(app)
		orderpb.RegisterOrderServiceServer(server, svc)
	})

	go server.RunHTTPServer(serviceName, func(router *gin.Engine) {
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
