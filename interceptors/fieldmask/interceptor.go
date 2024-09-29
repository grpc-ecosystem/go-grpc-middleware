package fieldmask

import (
	"context"

	"github.com/mennanov/fmutils"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var DefaultFilterFunc FilterFunc = func(msg proto.Message, paths []string) {
	fmutils.Filter(msg, paths)
}

type FilterFunc func(msg proto.Message, paths []string)

// UnaryServerInterceptor returns a new unary server interceptor that will decide whether to which fields should return to clients.
func UnaryServerInterceptor(filterFunc FilterFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		resp, err = handler(ctx, req)
		if err != nil {
			return
		}
		reqWithFieldMask, ok := req.(interface {
			GetFieldMask() *fieldmaskpb.FieldMask
		})
		if !ok {
			return
		}
		if len(reqWithFieldMask.GetFieldMask().GetPaths()) > 0 {
			protoResp, ok := resp.(proto.Message)
			if !ok {
				return
			}
			filterFunc(protoResp, reqWithFieldMask.GetFieldMask().GetPaths())
		}
		return
	}
}
