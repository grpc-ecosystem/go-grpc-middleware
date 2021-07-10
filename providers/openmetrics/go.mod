module github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2

go 1.15

require (
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-rc.2.0.20210128111500-3ff779b52992
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/client_model v0.2.0
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.37.0
	google.golang.org/protobuf v1.26.0
)

replace github.com/grpc-ecosystem/go-grpc-middleware/v2 => ../..
