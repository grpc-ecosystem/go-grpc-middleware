module github.com/grpc-ecosystem/go-grpc-middleware/interceptors/logging/examples

go 1.21

require (
	github.com/go-kit/log v0.2.1
	github.com/go-logr/logr v1.2.4
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0
	github.com/rs/zerolog v1.29.0
	github.com/sirupsen/logrus v1.9.0
	go.uber.org/zap v1.24.0
	google.golang.org/grpc v1.61.1
	k8s.io/klog/v2 v2.90.1
)

require (
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240213162025-012b6fc9bca9 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)

replace github.com/grpc-ecosystem/go-grpc-middleware/v2 => ../../../
