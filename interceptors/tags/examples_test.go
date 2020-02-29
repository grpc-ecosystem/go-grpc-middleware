package ctxtags_test

import (
	ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/interceptors/tags"
	"google.golang.org/grpc"
)

// Simple example of server initialization code, with data automatically populated from `log_fields` Golang tags.
func Example_initialization() {
	opts := []ctxtags.Option{
		ctxtags.WithFieldExtractor(ctxtags.TagBasedRequestFieldExtractor("log_fields")),
	}
	_ = grpc.NewServer(
		grpc.StreamInterceptor(ctxtags.StreamServerInterceptor(opts...)),
		grpc.UnaryInterceptor(ctxtags.UnaryServerInterceptor(opts...)),
	)
}
