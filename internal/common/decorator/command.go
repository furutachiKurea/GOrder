package decorator

import (
	"context"

	"github.com/rs/zerolog"
)

// CommandHandler 定义了一个泛型处理器，接收 Query C 并返回 result R
type CommandHandler[C, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

// ApplyCommandDecorators 为给定的 CommandHandler 应用装饰器，并返回装饰后的 handler
func ApplyCommandDecorators[C, R any](handler CommandHandler[C, R], logger zerolog.Logger, metricsClient MetricsClient) CommandHandler[C, R] {
	// queryLoggingDecorator, queryMetricsDecorator 均实现了 QueryHandler 接口，故可实现对于 handler 的装饰
	// 装饰器的顺序决定了调用链的顺序，最外层的装饰器最先被调用, defer 最后被调用
	return commandLoggingDecorator[C, R]{
		logger: logger,
		base: commandMetricsDecorator[C, R]{
			base:   handler,
			client: metricsClient,
		},
	}
}
