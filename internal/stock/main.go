package main

import (
	"log"

	"github.com/furutachiKurea/gorder/internal/common/config"
	"github.com/furutachiKurea/gorder/internal/common/genproto/stockpb"
	"github.com/furutachiKurea/gorder/internal/common/server"
	"github.com/furutachiKurea/gorder/internal/stock/ports"

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

	switch serverType {
	case "grpc":
		server.RunGRPCServer(serviceName, func(server *grpc.Server) {
			svc := ports.NewGRPCServer()
			stockpb.RegisterStockServiceServer(server, svc)
		})
	case "http":
		// TODO implement
	default:
		panic("unsupported service type: " + serverType)
	}
}
