package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// 初始化 TracerProvider 并注册为全局 tracer
func initTracer(jaegerEndpoint string) (func(context.Context) error, error) {
	// 创建 Jaeger exporter（指定 Jaeger 收集器地址）
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		return nil, err
	}

	// 配置资源（服务名、环境等元数据）
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("user-service"),   // 服务名称（关键，用于链路区分）
			semconv.DeploymentEnvironment("prod"), // 环境标识
		),
	)
	if err != nil {
		return nil, err
	}

	// 配置采样率（如 1.0 表示全量采样，生产环境建议 0.1）
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp), // 批量导出追踪数据（减少性能开销）
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(1))), // 10% 采样
	)

	// 将 tracerProvider 设置为全局
	otel.SetTracerProvider(tracerProvider)

	// 返回关闭函数（程序退出时调用，确保数据 flush）
	return tracerProvider.Shutdown, nil
}
