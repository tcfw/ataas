package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

//PassthroughMD creates a new outgoing ctx with the MD of the incoming cts
func PassthroughMD(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	ogCtx := metadata.NewOutgoingContext(ctx, md)
	return ogCtx
}
