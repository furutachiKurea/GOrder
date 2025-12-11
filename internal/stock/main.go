package main

import (
	"context"

	_ "github.com/furutachiKurea/gorder/common/config"
	"github.com/furutachiKurea/gorder/common/discovery"
	"github.com/furutachiKurea/gorder/common/genproto/stockpb"
	"github.com/furutachiKurea/gorder/common/logging"
	"github.com/furutachiKurea/gorder/common/server"
	"github.com/furutachiKurea/gorder/common/tracing"
	"github.com/furutachiKurea/gorder/stock/ports"
	"github.com/furutachiKurea/gorder/stock/service"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	logging.Init()
}

func main() {
	serviceName := viper.GetString("stock.service-name")
	serverType := viper.GetString("stock.server-to-run")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := tracing.InitJaegerProvider(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init jaeger provider")
	}
	defer func() {
		_ = shutdown(ctx)
	}()

	app := service.NewApplication(ctx)

	deregisterFn, err := discovery.RegisterToConsul(ctx, serviceName)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to register service %s to consul", serviceName)
	}
	defer func() { _ = deregisterFn() }()

	switch serverType {
	case "grpc":
		server.RunGRPCServer(serviceName, func(server *grpc.Server) {
			svc := ports.NewGRPCServer(app)
			stockpb.RegisterStockServiceServer(server, svc)
		})
	case "http":
		log.Panic().Str("serverType", serverType).Msg("unsupported server type")
	default:
		log.Panic().Str("serverType", serverType).Msg("unsupported server type")
	}
}
