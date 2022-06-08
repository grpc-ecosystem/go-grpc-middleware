module github.com/grpc-ecosystem/go-grpc-middleware/providers/opentelemetry/v2

go 1.14

replace github.com/grpc-ecosystem/go-grpc-middleware/v2 => ../..

require (
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-rc.2.0.20201002093600-73cf2ae9d891
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.32.0
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/sdk v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
	google.golang.org/grpc v1.46.2
)
