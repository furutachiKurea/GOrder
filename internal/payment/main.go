package main

import (
	"context"

	"github.com/furutachiKurea/gorder/common/config"
	"github.com/furutachiKurea/gorder/common/discovery"
	"github.com/furutachiKurea/gorder/common/logging"
	"github.com/furutachiKurea/gorder/common/server"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
	if err := config.NewViperConfig(); err != nil {
		logrus.Fatal(err)
	}
}
func main() {
	serviceName := viper.GetString("payment.service-name")
	serverType := viper.GetString("payment.server-to-run")

	paymentHandler := NewPaymentHandler()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deregisterFn, err := discovery.RegisterToConsul(ctx, serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() { _ = deregisterFn() }()

	switch serverType {
	case "grpc":
		logrus.Panic("unsupported type: ", serverType)
	case "http":
		server.RunHTTPServer(serviceName, paymentHandler.RegisterRoutes)
	default:
		logrus.Panic("unsupported service type: " + serverType)
	}
}
