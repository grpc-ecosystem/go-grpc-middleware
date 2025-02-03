module github.com/grpc-ecosystem/go-grpc-middleware/interceptors/logging/examples

go 1.21.1

toolchain go1.21.13

require (
	github.com/go-kit/log v0.2.1
	github.com/go-logr/logr v1.2.4
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0
	github.com/rs/zerolog v1.29.0
	github.com/sirupsen/logrus v1.9.0
	go.uber.org/zap v1.24.0
	google.golang.org/grpc v1.67.1
	k8s.io/klog/v2 v2.90.1
)

require (
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240826202546-f6391c0de4c7 // indirect
	google.golang.org/protobuf v1.36.4 // indirect
)

replace github.com/grpc-ecosystem/go-grpc-middleware/v2 => ../../../
