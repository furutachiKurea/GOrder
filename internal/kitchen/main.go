package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/furutachiKurea/gorder/common/broker"
	"github.com/furutachiKurea/gorder/common/client"
	_ "github.com/furutachiKurea/gorder/common/config"
	"github.com/furutachiKurea/gorder/common/logging"
	"github.com/furutachiKurea/gorder/common/tracing"
	"github.com/furutachiKurea/gorder/kitchen/adapter"
	"github.com/furutachiKurea/gorder/kitchen/infrastructure/consumer"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
}

func main() {
	serviceName := viper.GetString("kitchen.service-name")
	// serverType := viper.GetString("stock.server-to-run")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := tracing.InitJaegerProvider(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init jaeger provider")
	}
	defer func() {
		_ = shutdown(ctx)
	}()

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

	orderClient, closeOrderClient, err := client.NewOrderGRPCClient(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create order grpc client")
	}
	defer func() {
		_ = closeOrderClient()
	}()

	orderGRPC := adapter.NewOderGRPC(orderClient)
	go consumer.NewConsumer(orderGRPC).Listen(ch)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	<-ctx.Done()
}
