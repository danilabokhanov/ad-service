package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"runtime/debug"
	"time"
)

func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	timer := time.Now()

	res, err := handler(ctx, req)

	log.Println("full method:", info.FullMethod, "timer:", time.Since(timer), "error:", err)
	return res, err
}

func RecoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (_ interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("Panic: `%s` %s", info.FullMethod,
				string(debug.Stack()))
			err = status.Error(codes.Internal, msg)
		}
	}()
	return handler(ctx, req)
}
