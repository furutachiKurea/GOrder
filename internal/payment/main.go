package main

import (
	"github.com/furutachiKurea/gorder/common/broker"
	"github.com/furutachiKurea/gorder/common/config"
	"github.com/furutachiKurea/gorder/common/logging"
	"github.com/furutachiKurea/gorder/common/server"
	"github.com/rs/zerolog/log"

	"github.com/spf13/viper"
)

func init() {
	logging.Init()
	if err := config.NewViperConfig(); err != nil {
		log.Fatal().Err(err).Msg("failed to init config")
	}
}
func main() {
	serviceName := viper.GetString("payment.service-name")
	serverType := viper.GetString("payment.server-to-run")

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

	paymentHandler := NewPaymentHandler()

	switch serverType {
	case "grpc":
		log.Panic().Str("serverType", serverType).Msg("unsupported server type")
	case "http":
		server.RunHTTPServer(serviceName, paymentHandler.RegisterRoutes)
	default:
		log.Panic().Str("serverType", serverType).Msg("unsupported server type")
	}
}
