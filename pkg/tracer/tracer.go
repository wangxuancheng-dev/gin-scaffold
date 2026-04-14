// Package tracer 初始化 OpenTelemetry OTLP HTTP 导出（可对接 Jaeger 等后端）。
package tracer

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"gin-scaffold/config"
)

var tp *sdktrace.TracerProvider

// Init 注册全局 TracerProvider；未启用时返回 nil 无副作用。
func Init(ctx context.Context, cfg *config.TraceConfig) (func(context.Context) error, error) {
	if cfg == nil || !cfg.Enabled || cfg.Endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("otlp exporter: %w", err)
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}
	tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	shutdown := func(c context.Context) error {
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()
		if tp == nil {
			return nil
		}
		return tp.Shutdown(ctx)
	}
	return shutdown, nil
}
