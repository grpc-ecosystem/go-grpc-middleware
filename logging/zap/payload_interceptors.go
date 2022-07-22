package grpc_zap

import (
	"bytes"
	"context"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

var (
	// JsonPbMarshaller is the marshaller used for serializing protobuf messages.
	// If needed, this variable can be reassigned with a different marshaller with the same Marshal() signature.
	JsonPbMarshaller grpc_logging.JsonPbMarshaler = &jsonpb.Marshaler{}
)

type payloadOptions struct {
	logLevel zapcore.Level
}

// defaultPayloadOptions set a InfoLevel as default
func defaultPayloadOptions() *payloadOptions {
	return &payloadOptions{logLevel: zapcore.InfoLevel}
}

type PayloadOption func(o *payloadOptions)

// WithPayloadLogLevel set a log level for payload writes.
func WithPayloadLogLevel(logLevel zapcore.Level) PayloadOption {
	return func(o *payloadOptions) {
		o.logLevel = logLevel
	}
}

// PayloadUnaryServerInterceptor returns a new unary server interceptors that logs the payloads of requests.
//
// This *only* works when placed *after* the `grpc_zap.UnaryServerInterceptor`. However, the logging can be done to a
// separate instance of the logger.
func PayloadUnaryServerInterceptor(logger *zap.Logger, decider grpc_logging.ServerPayloadLoggingDecider, opts ...PayloadOption) grpc.UnaryServerInterceptor {
	options := defaultPayloadOptions()
	for _, opt := range opts {
		opt(options)
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		if !decider(ctx, info.FullMethod, info.Server) {
			return handler(ctx, req)
		}
		// Use the provided zap.Logger for logging but use the fields from context.
		logEntry := logger.With(append(serverCallFields(info.FullMethod), ctxzap.TagsToFields(ctx)...)...)
		logProtoMessageAsJson(logEntry, options.logLevel, req, "grpc.request.content", "server request payload logged as grpc.request.content field")
		resp, err := handler(ctx, req)
		if err == nil {
			logProtoMessageAsJson(logEntry, options.logLevel, resp, "grpc.response.content", "server response payload logged as grpc.response.content field")
		}
		return resp, err
	}
}

// PayloadStreamServerInterceptor returns a new server server interceptors that logs the payloads of requests.
//
// This *only* works when placed *after* the `grpc_zap.StreamServerInterceptor`. However, the logging can be done to a
// separate instance of the logger.
func PayloadStreamServerInterceptor(logger *zap.Logger, decider grpc_logging.ServerPayloadLoggingDecider, opts ...PayloadOption) grpc.StreamServerInterceptor {
	options := defaultPayloadOptions()
	for _, opt := range opts {
		opt(options)
	}
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !decider(stream.Context(), info.FullMethod, srv) {
			return handler(srv, stream)
		}
		logEntry := logger.With(append(serverCallFields(info.FullMethod), ctxzap.TagsToFields(stream.Context())...)...)
		newStream := &loggingServerStream{ServerStream: stream, logger: logEntry, logLevel: options.logLevel}
		return handler(srv, newStream)
	}
}

// PayloadUnaryClientInterceptor returns a new unary client interceptor that logs the payloads of requests and responses.
func PayloadUnaryClientInterceptor(logger *zap.Logger, decider grpc_logging.ClientPayloadLoggingDecider, opts ...PayloadOption) grpc.UnaryClientInterceptor {
	options := defaultPayloadOptions()
	for _, opt := range opts {
		opt(options)
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !decider(ctx, method) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		logEntry := logger.With(newClientLoggerFields(ctx, method)...)
		logProtoMessageAsJson(logEntry, options.logLevel, req, "grpc.request.content", "client request payload logged as grpc.request.content")
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			logProtoMessageAsJson(logEntry, options.logLevel, reply, "grpc.response.content", "client response payload logged as grpc.response.content")
		}
		return err
	}
}

// PayloadStreamClientInterceptor returns a new streaming client interceptor that logs the payloads of requests and responses.
func PayloadStreamClientInterceptor(logger *zap.Logger, decider grpc_logging.ClientPayloadLoggingDecider, opts ...PayloadOption) grpc.StreamClientInterceptor {
	options := defaultPayloadOptions()
	for _, opt := range opts {
		opt(options)
	}
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if !decider(ctx, method) {
			return streamer(ctx, desc, cc, method, opts...)
		}
		logEntry := logger.With(newClientLoggerFields(ctx, method)...)
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		newStream := &loggingClientStream{ClientStream: clientStream, logger: logEntry, logLevel: options.logLevel}
		return newStream, err
	}
}

type loggingClientStream struct {
	grpc.ClientStream
	logger   *zap.Logger
	logLevel zapcore.Level
}

func (l *loggingClientStream) SendMsg(m interface{}) error {
	err := l.ClientStream.SendMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.logger, l.logLevel, m, "grpc.request.content", "server request payload logged as grpc.request.content field")
	}
	return err
}

func (l *loggingClientStream) RecvMsg(m interface{}) error {
	err := l.ClientStream.RecvMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.logger, l.logLevel, m, "grpc.response.content", "server response payload logged as grpc.response.content field")
	}
	return err
}

type loggingServerStream struct {
	grpc.ServerStream
	logger   *zap.Logger
	logLevel zapcore.Level
}

func (l *loggingServerStream) SendMsg(m interface{}) error {
	err := l.ServerStream.SendMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.logger, l.logLevel, m, "grpc.response.content", "server response payload logged as grpc.response.content field")
	}
	return err
}

func (l *loggingServerStream) RecvMsg(m interface{}) error {
	err := l.ServerStream.RecvMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.logger, l.logLevel, m, "grpc.request.content", "server request payload logged as grpc.request.content field")
	}
	return err
}

func logProtoMessageAsJson(logger *zap.Logger, logLevel zapcore.Level, pbMsg interface{}, key string, msg string) {
	if p, ok := pbMsg.(proto.Message); ok {
		// logger.Check may return `nil`
		if ce := logger.Check(logLevel, msg); ce != nil {
			ce.Write(zap.Object(key, &jsonpbObjectMarshaler{pb: p}))
		}
	}
}

type jsonpbObjectMarshaler struct {
	pb proto.Message
}

func (j *jsonpbObjectMarshaler) MarshalLogObject(e zapcore.ObjectEncoder) error {
	// ZAP jsonEncoder deals with AddReflect by using json.MarshalObject. The same thing applies for consoleEncoder.
	return e.AddReflected("msg", j)
}

func (j *jsonpbObjectMarshaler) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := JsonPbMarshaller.Marshal(b, j.pb); err != nil {
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}
	return b.Bytes(), nil
}
