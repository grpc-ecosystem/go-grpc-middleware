package tags_test

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
	"google.golang.org/grpc"
)

// Simple example of server initialization code, with data automatically populated from `log_fields` Golang tags.
func Example_initialization() {
	opts := []tags.Option{
		tags.WithFieldExtractor(tags.TagBasedRequestFieldExtractor("log_fields")),
	}
	_ = grpc.NewServer(
		grpc.StreamInterceptor(tags.StreamServerInterceptor(opts...)),
		grpc.UnaryInterceptor(tags.UnaryServerInterceptor(opts...)),
	)
}
