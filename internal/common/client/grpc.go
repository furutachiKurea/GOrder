package client

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/furutachiKurea/gorder/common/discovery"
	"github.com/furutachiKurea/gorder/common/genproto/orderpb"
	"github.com/furutachiKurea/gorder/common/genproto/stockpb"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewStockGRPCClient(ctx context.Context) (
	client stockpb.StockServiceClient,
	close func() error,
	err error,
) {
	if !waitForStockGRPCClient(viper.GetDuration("dial-grpc-timeout")) {
		return nil, func() error { return nil }, errors.New("stock grpc not available")
	}

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
	if !waitForOrderGRPCClient(viper.GetDuration("dial-grpc-timeout")) {
		return nil, func() error { return nil }, errors.New("order grpc not available")
	}

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
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
}

func waitForStockGRPCClient(timeout time.Duration) bool {
	log.Info().Str("timeout", timeout.String()).Msg("waiting for stock grpc client")
	return waitFor(viper.GetString("stock.grpc-addr"), timeout)
}

func waitForOrderGRPCClient(timeout time.Duration) bool {
	log.Info().Str("timeout", timeout.String()).Msg("waiting for order grpc client")
	return waitFor(viper.GetString("order.grpc-addr"), timeout)
}

// waitFor 尝试连接 addr 直至 timeout
func waitFor(addr string, timeout time.Duration) bool {
	portAvailable := make(chan struct{})
	timeoutCh := time.After(timeout)

	go func() {
		for {
			select {
			case <-timeoutCh:
				return
			default:
				// continue
			}

			_, err := net.DialTimeout("tcp", addr, timeout)
			if err == nil {
				close(portAvailable)
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	select {
	case <-portAvailable:
		return true
	case <-timeoutCh:
		return false
	}
}
