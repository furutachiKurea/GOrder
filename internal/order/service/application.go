package service

import (
	"context"
	"fmt"

	"github.com/furutachiKurea/gorder/common/broker"
	grpcclient "github.com/furutachiKurea/gorder/common/client"
	"github.com/furutachiKurea/gorder/common/metrics"
	"github.com/furutachiKurea/gorder/order/adapter"
	"github.com/furutachiKurea/gorder/order/adapter/grpc"
	"github.com/furutachiKurea/gorder/order/app"
	"github.com/furutachiKurea/gorder/order/app/command"
	"github.com/furutachiKurea/gorder/order/app/query"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func NewApplication(ctx context.Context) (app app.Application, close func()) {
	stockClient, closeStockClient, err := grpcclient.NewStockGRPCClient(ctx)
	if err != nil {
		panic(err)
	}
	stockGRPC := grpc.NewStockGRPC(stockClient)

	ch, closeCoon := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)
	mongoClient, disconnectMongo := newMongoClient(ctx)
	return newApplication(ctx, stockGRPC, mongoClient, ch), func() {
		_ = closeStockClient()
		_ = ch.Close()
		_ = closeCoon()
		_ = disconnectMongo(ctx)
	}

}

func newApplication(ctx context.Context, stockClient query.StockInterface, mongoClient *mongo.Client, ch *amqp.Channel) app.Application {
	orderRepo := adapter.NewOrderRepositoryMongo(mongoClient)
	logger := log.Logger
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreateOrder: command.NewCreateOrderHandler(
				orderRepo,
				stockClient,
				ch,
				logger,
				metricsClient,
			),
			UpdateOrder: command.NewUpdateOrderHandler(
				orderRepo,
				logger,
				metricsClient,
			),
		},
		Queries: app.Queries{
			GetCustomerOrder: query.NewGetCustomerOrderHandler(
				orderRepo,
				logger,
				metricsClient,
			),
		},
	}

}

func newMongoClient(ctx context.Context) (*mongo.Client, func(ctx context.Context) error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s",
		viper.GetString("mongo.user"),
		viper.GetString("mongo.password"),
		viper.GetString("mongo.host"),
		viper.GetString("mongo.port"),
	)

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	if err = c.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	return c, c.Disconnect
}
