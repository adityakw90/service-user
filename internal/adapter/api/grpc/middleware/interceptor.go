package middleware

import (
	"context"

	"google.golang.org/grpc"
)

// ChainUnaryInterceptors allows you to chain multiple unary interceptors into one.
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	n := len(interceptors)

	if n == 0 {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	if n == 1 {
		return interceptors[0]
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		currHandler := handler
		for i := n - 1; i >= 0; i-- {
			currInterceptor := interceptors[i]
			currHandler = func(currCtx context.Context, currReq interface{}) (interface{}, error) {
				return currInterceptor(currCtx, currReq, info, currHandler)
			}
		}
		return currHandler(ctx, req)
	}
}

// ChainStreamInterceptors allows you to chain multiple stream interceptors into one.
func ChainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	n := len(interceptors)

	if n == 0 {
		return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return handler(srv, ss)
		}
	}

	if n == 1 {
		return interceptors[0]
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		currHandler := handler
		for i := n - 1; i >= 0; i-- {
			currInterceptor := interceptors[i]
			currHandler = func(currSrv interface{}, currSS grpc.ServerStream) error {
				return currInterceptor(currSrv, currSS, info, currHandler)
			}
		}
		return currHandler(srv, ss)
	}
}
