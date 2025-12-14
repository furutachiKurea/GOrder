package logging

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TODO return 的 logger 无法实际的被返回的函数使用

func WhenCommandExecute(ctx context.Context, commandName string, cmd any, err error) {
	l := log.With().Any("cmd", cmd).Logger()
	if err == nil {
		l.Info().Ctx(ctx).Msgf("%s_command_success", commandName)
	} else {
		l.Error().Ctx(ctx).Msgf("%s_command_failed", commandName)
	}
}

func WhenRequest(ctx context.Context, method string, args ...any) (zerolog.Logger, func(resp any, errP *error)) {
	l := log.With().Str(Method, method).Str(Args, formatArgs(args)).Logger()

	start := time.Now()
	return l, func(resp any, err *error) {
		l = l.With().Int(Cost, int(time.Since(start).Milliseconds())).
			Any(Response, resp).Logger()

		if err != nil && *err != nil {
			l.Error().
				Ctx(ctx).
				Str(Error, (*err).Error()).
				Msgf("%s_request_failed", method)
		}

		l.Info().Ctx(ctx).Msgf("%s_request_success", method)
	}
}

func WhenEventPublish(ctx context.Context, args ...any) (zerolog.Logger, func(any, *error)) {
	l := log.With().
		Str(Args, formatArgs(args)).
		Logger()

	start := time.Now()
	return l, func(resp any, err *error) {
		l = l.With().Int(Cost, int(time.Since(start).Milliseconds())).
			Any(Response, resp).Logger()

		if err != nil && *err != nil {
			l.Error().
				Ctx(ctx).
				Str(Error, (*err).Error()).
				Msg("_mq_publish_failed")
		}

		l.Info().Ctx(ctx).Msg("_mq_publish_success")
	}
}
