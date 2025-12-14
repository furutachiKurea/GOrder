package server

import (
	"context"
	"fmt"
	"net"

	"github.com/furutachiKurea/gorder/common/logging"
	grpctags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpclogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func RunGRPCServer(serviceName string, registerServer func(server *grpc.Server)) {
	addr := viper.Sub(serviceName).GetString("grpc-addr")
	if addr == "" {
		log.Warn().Msg("grpc-addr is empty, use fallback-grpc-addr instead")
		addr = viper.GetString("fallback-grpc-addr")
	}

	RunGRPCServerOnAddr(addr, registerServer)
}

func RunGRPCServerOnAddr(addr string, registerServer func(server *grpc.Server)) {
	logger := log.Logger

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		// 相当于中间件
		grpc.ChainUnaryInterceptor(
			grpctags.UnaryServerInterceptor(grpctags.WithFieldExtractor(grpctags.CodeGenRequestFieldExtractor)),
			grpclogging.UnaryServerInterceptor(InterceptorLogger(logger)),
			logging.GRPCUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			grpctags.StreamServerInterceptor(grpctags.WithFieldExtractor(grpctags.CodeGenRequestFieldExtractor)),
			grpclogging.StreamServerInterceptor(InterceptorLogger(logger)),
		), // 拦截流式调用
	)

	registerServer(grpcServer)

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panic().Err(err).Msg("")
	}

	log.Info().Msgf("Starting gRPC server on %s", addr)
	if err := grpcServer.Serve(listen); err != nil {
		log.Panic().Err(err).Msg("")
	}
}

// InterceptorLogger adapts zerolog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l zerolog.Logger) grpclogging.Logger {
	return grpclogging.LoggerFunc(func(ctx context.Context, lvl grpclogging.Level, msg string, fields ...any) {
		l := l.With().Fields(fields).Logger()

		switch lvl {
		case grpclogging.LevelDebug:
			l.Debug().Msg(msg)
		case grpclogging.LevelInfo:
			l.Info().Msg(msg)
		case grpclogging.LevelWarn:
			l.Warn().Msg(msg)
		case grpclogging.LevelError:
			l.Error().Msg(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
