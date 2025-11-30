package decorator

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

type queryLoggingDecorator[C, R any] struct {
	logger zerolog.Logger
	base   QueryHandler[C, R]
}

func (q queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	logger := q.logger.With().
		Str("query", generateActionName(cmd)).
		Str("query_body", fmt.Sprintf("%#v", cmd)).
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

func generateActionName(cmd any) string {
	return strings.Split(fmt.Sprintf("%T", cmd), ".")[1]
}
