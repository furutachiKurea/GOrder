package decorator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

type queryLoggingDecorator[C, R any] struct {
	logger zerolog.Logger
	base   QueryHandler[C, R]
}

func (q queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	body, _ := json.Marshal(cmd)
	logger := q.logger.With().
		Str("query", generateActionName(cmd)).
		Str("query_body", string(body)).
		Logger()

	logger.Debug().Msg("Executing query")
	defer func() {
		if err == nil {
			logger.Info().Msg("Query executed successfully")
		} else {
			logger.Error().Err(err).Msg("Failed to execute query")
		}
	}()

	return q.base.Handle(ctx, cmd)
}

type commandLoggingDecorator[C, R any] struct {
	logger zerolog.Logger
	base   QueryHandler[C, R]
}

func (r commandLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	body, _ := json.Marshal(cmd)
	logger := r.logger.With().
		Str("command", generateActionName(cmd)).
		Str("command_body", string(body)).
		Logger()

	defer func() {
		if err == nil {
			logger.Info().Msg("Command executed successfully")
		} else {
			logger.Error().Err(err).Msg("Failed to execute Command")
		}
	}()

	return r.base.Handle(ctx, cmd)
}

func generateActionName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
