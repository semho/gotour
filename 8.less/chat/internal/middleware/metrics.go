package middleware

import (
	"chat/pkg/metrics"
	"context"
	"time"

	"google.golang.org/grpc"
)

func MetricsInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	//Rate
	metrics.RequestsTotal.WithLabelValues(info.FullMethod).Inc()

	resp, err := handler(ctx, req)

	duration := time.Since(start).Seconds()
	//Duration
	metrics.RequestDuration.WithLabelValues(info.FullMethod).Observe(duration)

	if err != nil {
		//Errors
		metrics.ErrorsTotal.WithLabelValues(info.FullMethod).Inc()
	}

	return resp, err
}
