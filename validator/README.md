# gRPC Validation Interceptors

Package `grpc_validator` provides an easy way to hook protobuf message validation as a gRPC
interceptor across all your APIs.

It primarily meant to be used with https://github.com/mwitkow/go-proto-validators, which code-gen
assertions about allowed values from `.proto` files.