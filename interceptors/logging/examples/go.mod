module github.com/grpc-ecosystem/go-grpc-middleware/interceptors/logging/examples

go 1.24.0

require (
	github.com/go-kit/log v0.2.1
	github.com/go-logr/logr v1.4.3
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0
	github.com/rs/zerolog v1.34.0
	github.com/sirupsen/logrus v1.9.3
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.76.0
	k8s.io/klog/v2 v2.130.1
)

require (
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/grpc-ecosystem/go-grpc-middleware/v2 => ../../../
