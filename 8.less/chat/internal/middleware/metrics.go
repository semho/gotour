package middleware

import (
	"chat/pkg/metrics"
	"context"
	"google.golang.org/grpc"
	"time"
)

func MetricsInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	metrics.RequestsTotal.WithLabelValues(info.FullMethod).Inc()

	resp, err := handler(ctx, req)

	duration := time.Since(start).Seconds()
	metrics.RequestDuration.WithLabelValues(info.FullMethod).Observe(duration)

	if err != nil {
		metrics.ErrorsTotal.WithLabelValues(info.FullMethod).Inc()
	}

	return resp, err
}
