package grpc

import (
	"context"
	"runtime"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

//InternalClientOptions to be used for GRPC clients to INTERNAL services
func InternalClientOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(&PassthroughPerRPCCreds{}),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			otelgrpc.UnaryClientInterceptor(),
			grpc_prometheus.UnaryClientInterceptor,
		)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			otelgrpc.StreamClientInterceptor(),
			grpc_prometheus.StreamClientInterceptor,
		)),
		grpc.WithDefaultCallOptions(grpc.UseCompressor("brotli")),
	}
}

//ExternalClientOptions to be used for GRPC clients to EXTERNAL services
func ExternalClientOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			otelgrpc.UnaryClientInterceptor(),
			grpc_prometheus.UnaryClientInterceptor,
		)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			otelgrpc.StreamClientInterceptor(),
			grpc_prometheus.StreamClientInterceptor,
		)),
		grpc.WithDefaultCallOptions(grpc.UseCompressor("brotli")),
	}
}

//DefaultServerOptions to be used for GRPC servers
func DefaultServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(recoveryFunc())),
			otelgrpc.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(),
			otelgrpc.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
		)),
	}
}

var (
	allowedAuthRefs = map[string]string{
		"authorization":    "authorization",
		"user-agent":       "user-agent",
		"x-forwarded-host": "x-forwarded-host",
		"x-forwarded-for":  "x-forwarded-for",
	}
)

type PassthroughPerRPCCreds struct{}

func (c *PassthroughPerRPCCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	outgoingMD := map[string]string{}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return outgoingMD, nil
	}

	for k, vs := range md {
		nk, ok := allowedAuthRefs[k]
		if ok {
			outgoingMD[nk] = vs[0]
		}
	}

	return outgoingMD, nil
}

func (c *PassthroughPerRPCCreds) RequireTransportSecurity() bool {
	return false
}

func recoveryFunc() grpc_recovery.RecoveryHandlerFunc {
	log := logrus.New()

	return func(p interface{}) (err error) {
		_, file, ln, _ := runtime.Caller(9)
		log.WithFields(logrus.Fields{
			"file": file,
			"line": ln,
		}).Errorf("%s", p)
		return status.Errorf(codes.Internal, "%s", p)
	}
}
