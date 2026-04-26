package middleware

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adityakw90/service-user/internal/adapter/api/grpc/response"
	domainErrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/pkg/util"

	monitoring "github.com/adityakw90/go-monitoring"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// WrappedServerStream is a wrapper around grpc.ServerStream that allows us to inject a custom context.
type WrappedServerStream struct {
	grpc.ServerStream
	wrappedCtx context.Context
}

// Context returns the wrapped context instead of the original.
func (w *WrappedServerStream) Context() context.Context {
	return w.wrappedCtx
}

// UnaryRequestInterceptor is a gRPC interceptor for logging incoming requests.
func UnaryRequestInterceptor(
	m *monitoring.Monitoring,
) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// initialize
		var err error
		var resp any

		// Extract metadata from incoming request
		clientName := "unknown"
		actorId := "unknown"
		actorType := "unknown"
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			ctx = m.Tracer.ExtractContext(ctx, md)

			if mdClient, exists := md["client"]; exists && len(mdClient) > 0 {
				clientName = mdClient[0]
			}
			if mdActorId, exists := md["actor-id"]; exists && len(mdActorId) > 0 {
				actorId = mdActorId[0]
			}
			if mdActorType, exists := md["actor-type"]; exists && len(mdActorType) > 0 {
				actorType = mdActorType[0]
			}
		}

		// Start a new trace span for the request
		ctx, span := m.Tracer.StartSpan(ctx, info.FullMethod)
		defer m.Tracer.EndSpan(span)

		// get logger with spancontext
		logger := m.Logger.WithSpanContext(span.SpanContext())

		// Extract traceID from the span context
		traceID := span.SpanContext().TraceID().String()

		// Store client name in context
		ctx = util.SetClientName(ctx, clientName)
		ctx = util.SetActor(ctx, actorId, actorType)

		// Add useful attributes to the span
		span.SetAttributes(
			attribute.String("service.type", "gRPC"),
			attribute.String("rpc.type", "unary"),
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", info.FullMethod),
			attribute.String("trace.id", traceID),
			attribute.String("client.name", clientName),
			attribute.String("actor.id", actorId),
			attribute.String("actor.type", actorType),
		)

		// start trace request
		span.AddEvent("gRPC Request Start", trace.WithAttributes(
			attribute.String("method", info.FullMethod),
		))

		// Start tracking request metrics
		requestCounter, err := m.Metric.CreateCounter("grpc_requests_total", "req", "Total number of gRPC requests")
		if err != nil {
			return nil, err
		}
		durationHistogram, err := m.Metric.CreateHistogram("grpc_request_duration_milliseconds", "ms", "Duration of gRPC requests in milliseconds")
		if err != nil {
			return nil, err
		}

		// Measure the start time
		startTime := time.Now()

		// Increment the request counter
		m.Metric.RecordCounter(ctx, requestCounter, 1, attribute.String("rpc.service", info.FullMethod))

		// Handle the request
		resp, err = handler(ctx, req)

		// Measure the request duration
		duration := time.Since(startTime).Milliseconds()
		m.Metric.RecordHistogram(ctx, durationHistogram, duration, attribute.String("rpc.service", info.FullMethod))

		// Log the response status
		if err != nil {
			grpcErr := response.MakeErrorResponse(err)
			code := status.Code(grpcErr).String()

			logData := map[string]interface{}{
				"rpc.Method":       info.FullMethod,
				"error.Type":       fmt.Sprintf("%T", err), // Captures Go's error type
				"error.Message":    err.Error(),            // original error message
				"response.Code":    code,
				"response.Message": grpcErr.Error(),
			}
			logAttr := []attribute.KeyValue{
				attribute.String("rpc.Method", info.FullMethod),
				attribute.String("error.Type", fmt.Sprintf("%T", err)),
				attribute.String("error.Message", err.Error()),
				attribute.String("response.Code", code),
				attribute.String("response.Message", grpcErr.Error()),
			}

			// Check if it's an unmapped error code (internal server error)
			if code == codes.Internal.String() {
				// Check for specific CustomError codes
				var customErr *domainErrors.CustomError
				if errors.As(err, &customErr) {
					switch customErr {
					case domainErrors.ErrInternalServerError:
						logData["Warning"] = "Unmapped API Error"
						logAttr = append(logAttr, attribute.String("Warning", "Unmapped API Error"))
					case domainErrors.ErrTraceInformationMissing:
						logData["Warning"] = "Trace information missing in request"
						logAttr = append(logAttr, attribute.String("Warning", "Trace information missing in request"))
					}
				} else {
					logData["Error"] = "Unhandled Build Response"
					logAttr = append(logAttr, attribute.String("Error", "Unhandled Build Response"))
				}
			}

			span.AddEvent("gRPC Request Failed", trace.WithAttributes(logAttr...))
			logger.Info("gRPC Request Failed", logData)
			return resp, grpcErr
		} else {
			span.AddEvent("gRPC Request Success", trace.WithAttributes(
				attribute.String("method", info.FullMethod),
			))
			logger.Info("gRPC Request Success", map[string]interface{}{
				"rpc.service": info.FullMethod,
			})
		}

		return resp, err
	}
}

func StreamRequestInterceptor(
	m *monitoring.Monitoring,
) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		var err error
		ctx := ss.Context()

		// Extract metadata from incoming request
		clientName := "unknown"
		actorId := "unknown"
		actorType := "unknown"
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			ctx = m.Tracer.ExtractContext(ctx, md)

			if mdClient, exists := md["client"]; exists && len(mdClient) > 0 {
				clientName = mdClient[0]
			}
			if mdActorId, exists := md["actor-id"]; exists && len(mdActorId) > 0 {
				actorId = mdActorId[0]
			}
			if mdActorType, exists := md["actor-type"]; exists && len(mdActorType) > 0 {
				actorType = mdActorType[0]
			}
		}

		// Start a new trace span for the streaming request
		ctx, span := m.Tracer.StartSpan(ctx, info.FullMethod)
		defer m.Tracer.EndSpan(span)

		// get logger with spancontext
		logger := m.Logger.WithSpanContext(span.SpanContext())

		// Extract traceID from the span context
		traceID := span.SpanContext().TraceID().String()

		// Store client name in context
		ctx = util.SetClientName(ctx, clientName)
		ctx = util.SetActor(ctx, actorId, actorType)

		// Add useful attributes to the span
		span.SetAttributes(
			attribute.String("service.type", "gRPC"),
			attribute.String("rpc.type", "stream"),
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", info.FullMethod),
			attribute.String("trace.id", traceID),
			attribute.String("client.name", clientName),
			attribute.String("actor.id", actorId),
			attribute.String("actor.type", actorType),
		)

		// start trace request
		span.AddEvent("gRPC Request Start", trace.WithAttributes(
			attribute.String("method", info.FullMethod),
		))

		// Start tracking request metrics
		requestCounter, err := m.Metric.CreateCounter("grpc_stream_requests_total", "req", "Total number of gRPC stream requests")
		if err != nil {
			return err
		}
		durationHistogram, err := m.Metric.CreateHistogram("grpc_stream_request_duration_milliseconds", "ms", "Duration of gRPC stream requests in milliseconds")
		if err != nil {
			return err
		}
		// Measure the start time
		startTime := time.Now()

		// Increment the request counter
		m.Metric.RecordCounter(ctx, requestCounter, 1, attribute.String("rpc.service", info.FullMethod))

		// Wrap the server stream with the new context
		wrappedStream := &WrappedServerStream{
			ServerStream: ss,
			wrappedCtx:   ctx,
		}

		// Handle the stream
		err = handler(srv, wrappedStream)

		// Measure the request duration
		duration := time.Since(startTime).Milliseconds()
		m.Metric.RecordHistogram(ctx, durationHistogram, duration, attribute.String("rpc.service", info.FullMethod))

		// Log the end of the stream
		if err != nil {
			grpcErr := response.MakeErrorResponse(err)
			code := status.Code(grpcErr).String()

			logData := map[string]interface{}{
				"rpc.Method":       info.FullMethod,
				"error.Type":       fmt.Sprintf("%T", err), // Captures Go's error type
				"error.Message":    err.Error(),            // original error message
				"response.Code":    code,
				"response.Message": grpcErr.Error(),
			}
			logAttr := []attribute.KeyValue{
				attribute.String("rpc.Method", info.FullMethod),
				attribute.String("error.Type", fmt.Sprintf("%T", err)),
				attribute.String("error.Message", err.Error()),
				attribute.String("response.Code", code),
				attribute.String("response.Message", grpcErr.Error()),
			}

			// Check if it's an unmapped error code (internal server error)
			if code == codes.Internal.String() {
				// Check for specific CustomError codes
				var customErr *domainErrors.CustomError
				if errors.As(err, &customErr) {
					switch customErr {
					case domainErrors.ErrInternalServerError:
						logData["Warning"] = "Unmapped API Error"
						logAttr = append(logAttr, attribute.String("Warning", "Unmapped API Error"))
					case domainErrors.ErrTraceInformationMissing:
						logData["Warning"] = "Trace information missing in request"
						logAttr = append(logAttr, attribute.String("Warning", "Trace information missing in request"))
					}
				} else {
					logData["Error"] = "Unhandled Build Response"
					logAttr = append(logAttr, attribute.String("Error", "Unhandled Build Response"))
				}
			}
			span.AddEvent("gRPC Request Failed", trace.WithAttributes(logAttr...))
			logger.Info("gRPC Request Failed", logData)

			return grpcErr
		} else {
			span.AddEvent("gRPC Request Success", trace.WithAttributes(
				attribute.String("method", info.FullMethod),
			))
			logger.Info("gRPC Request Success", map[string]interface{}{
				"rpc.service": info.FullMethod,
			})
		}

		return err
	}
}
