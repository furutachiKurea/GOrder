package client

import (
	"context"

	"github.com/furutachiKurea/gorder/common/genproto/stockpb"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func NewStockGRPCClient(_ context.Context) (
	client stockpb.StockServiceClient,
	close func() error,
	err error,
) {
	grpcAddr := viper.GetString("stock.grpc-addr")
	opts, err := grpcDialOpts()
	if err != nil {
		return nil, func() error { return nil }, err
	}

	coon, err := grpc.NewClient(grpcAddr, opts...)
	if err != nil {
		return nil, func() error { return nil }, err
	}

	return stockpb.NewStockServiceClient(coon), coon.Close, nil
}

func grpcDialOpts() ([]grpc.DialOption, error) {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, nil
}
