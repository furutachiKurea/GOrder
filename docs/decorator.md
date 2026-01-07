# decorator-装饰器

> 在 Go 里，装饰器（Decorator）不是语言特性，而是一种设计模式。
它的核心思想是：<br>
> 在不修改原有对象的前提下，用包装（wrap）的方式给它增加额外功能

一般包装接口、函数

```go
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
```

装饰器：

```go
type queryLoggingDecorator[C, R any] struct {
	logger *logrus.Entry
	base   QueryHandler[C, R]
}

func (q queryLoggingDecorator[C, R]) Handle(ctx context.Context, cmd C) (result R, err error) {
	logger := q.logger.WithFields(logrus.Fields{
		"query":      generateActionName(cmd),
		"query_body": fmt.Sprintf("%#v", cmd),
	})

	logger.Debug("Executing query")
	defer func() {
		if err == nil {
			logger.Info("Query executed successfully")
		} else {
			logger.Error("Failed to execute query: ", err)
		}
	}()
	return q.base.Handle(ctx, cmd)
}
```
