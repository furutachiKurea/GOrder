package decorator

import (
	"context"

	"github.com/sirupsen/logrus"
)

// QueryHandler 定义了一个泛型处理器，接收 Query Q 并返回 result R
type QueryHandler[Q, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

// ApplyQueryDecorators 为给定的 QueryHandler 应用装饰器，并返回装饰后的 handler
func ApplyQueryDecorators[H, R any](handler QueryHandler[H, R], logger *logrus.Entry, metricsClient MetricsClient) QueryHandler[H, R] {
	// queryLoggingDecorator, queryMetricsDecorator 均实现了 QueryHandler 接口，故可实现对于 handler 的装饰
	// 装饰器的顺序决定了调用链的顺序，最外层的装饰器最先被调用, defer 最后被调用
	return queryLoggingDecorator[H, R]{
		logger: logger,
		base: queryMetricsDecorator[H, R]{
			base:   handler,
			client: metricsClient,
		},
	}
}
