package tracing

import (
	"context"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("default_tracer")

// InitJaegerProvider 初始化 Jaeger Tracer Provider 并返回 providerShutdown 函数
func InitJaegerProvider(jaegerURL, serviceName string) (providerShutdown func(ctx context.Context) error, err error) {
	if jaegerURL == "" {
		panic("empty jaeger url")
	}

	tracer = otel.Tracer(serviceName)
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerURL)))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)
	b3Progators := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))
	p := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}, b3Progators,
	)

	otel.SetTextMapPropagator(p)

	return tp.Shutdown, nil
}

// Start 对 tracer.Start 的封装，使用组件级 tracer
//
// tracer.Start 会创建一个 span 和一个包含新创建 span 的 context.Context。
// 如果 `ctx` 中提供的 context.Context 包含一个 Span，那么新创建的 Span 将是该 span 的子 span；否则，它将是一个根 span。可以通过将 `WithNewRoot()` 作为 SpanOption 提供来覆盖此行为，即使 `ctx` 包含 Span，新创建的 Span 也会成为根 span。
// 创建 Span 时，建议使用 `WithAttributes()` SpanOption 提供所有已知的 span 属性，因为采样器只能访问创建 Span 时提供的属性。
// 任何创建的 Span 都必须被结束。这是用户的责任。如果 Span 未被结束，此 API 的实现可能会泄漏内存或其他资源。
func Start(ctx context.Context, name string) (context.Context, trace.Span) {
	return tracer.Start(ctx, name)
}

func TraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	return spanCtx.TraceID().String()
}
