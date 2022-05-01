module github.com/grpc-ecosystem/go-grpc-middleware/providers/opentracing/v2

go 1.14

replace github.com/grpc-ecosystem/go-grpc-middleware/v2 => ../..

require (
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-rc.2.0.20201002093600-73cf2ae9d891
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.37.0
)
