package client

import (
	"context"

	"github.com/furutachiKurea/gorder/common/discovery"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/common/genproto/stockpb"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewStockGRPCClient(ctx context.Context) (
	client stockpb.StockServiceClient,
	close func() error,
	err error,
) {
	grpcAddr, err := discovery.GetServiceAddr(ctx, viper.GetString("stock.service-name"))
	if err != nil {
		return nil, func() error { return nil }, err
	}

	opts := grpcDialOpts()

	coon, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error { return nil }, err
	}

	return stockpb.NewStockServiceClient(coon), coon.Close, nil
}

func NewOrderGRPCClient(ctx context.Context) (
	client orderpb.OrderServiceClient,
	close func() error,
	err error,
) {
	grpcAddr, err := discovery.GetServiceAddr(ctx, viper.GetString("order.service-name"))
	if err != nil {
		return nil, func() error { return nil }, err
	}

	opts := grpcDialOpts()

	coon, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error { return nil }, err
	}

	return orderpb.NewOrderServiceClient(coon), coon.Close, nil
}

func grpcDialOpts() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
}
