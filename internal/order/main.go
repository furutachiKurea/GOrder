package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/furutachiKurea/gorder/internal/common/config"
	"github.com/furutachiKurea/gorder/internal/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/internal/common/server"
	"github.com/furutachiKurea/gorder/internal/order/ports"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	if err := config.NewViperConfig(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	serviceName := viper.GetString("order.service-name")

	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		svc := ports.NewGRPCServer()
		orderpb.RegisterOrderServiceServer(server, svc)
	})

	go server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		ports.RegisterHandlersWithOptions(router, HTTPServer{}, ports.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	<-ctx.Done()

}
