package logging

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func GRPCUnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	fields := map[string]any{
		Args: req,
	}

	defer func() {
		fields[Response] = resp
		if err != nil {
			fields[Error] = err.Error()
			log.Error().Ctx(ctx).Fields(fields).Msg("_grpc_request_out")
		}
	}()

	md, exists := metadata.FromIncomingContext(ctx)
	if exists {
		fields["grpc_metadata"] = md
	}

	log.Info().Ctx(ctx).Fields(fields).Msg("_grpc_request_in")
	return handler(ctx, req)
}
