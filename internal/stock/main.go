package main

import (
	"context"
	"log"

	"github.com/furutachiKurea/gorder/common/config"
	"github.com/furutachiKurea/gorder/common/genproto/stockpb"
	"github.com/furutachiKurea/gorder/common/server"
	"github.com/furutachiKurea/gorder/stock/ports"
	"github.com/furutachiKurea/gorder/stock/service"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	if err := config.NewViperConfig(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	serviceName := viper.GetString("stock.service-name")
	serverType := viper.GetString("stock.server-to-run")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := service.NewApplication(ctx)

	switch serverType {
	case "grpc":
		server.RunGRPCServer(serviceName, func(server *grpc.Server) {
			svc := ports.NewGRPCServer(app)
			stockpb.RegisterStockServiceServer(server, svc)
		})
	case "http":
		// TODO implement
	default:
		panic("unsupported service type: " + serverType)
	}
}
