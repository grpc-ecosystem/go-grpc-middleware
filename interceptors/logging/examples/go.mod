module github.com/grpc-ecosystem/go-grpc-middleware/interceptors/logging/examples

go 1.19

require (
	github.com/go-kit/log v0.2.1
	github.com/go-logr/logr v1.2.4
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0
	github.com/rs/zerolog v1.29.0
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.0
	go.uber.org/zap v1.24.0
	golang.org/x/exp v0.0.0-20230321023759-10a507213a29
	google.golang.org/grpc v1.53.0
	k8s.io/klog/v2 v2.90.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/grpc-ecosystem/go-grpc-middleware/v2 => ../../../
