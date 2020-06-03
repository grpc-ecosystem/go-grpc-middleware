package zerolog

import (
	"bytes"
	"context"
	"fmt"

	//nolint:staticcheck  // Proto v1 is deprecated; v2 doesn't work yet.
	"github.com/golang/protobuf/jsonpb"
	//nolint:staticcheck  // Proto v1 is deprecated; v2 doesn't work yet.
	"github.com/golang/protobuf/proto"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	"github.com/irridia/go-grpc-middleware/logging/zerolog/ctxzerolog"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

var (
	// JSONPbMarshaller is the marshaller used for serializing protobuf messages.
	// If needed, this variable can be reassigned with a different marshaller with the same Marshal() signature.
	// NOTE: We need to maintain proto v1 support since the test Messages are still v1.
	JSONPbMarshaller grpc_logging.JsonPbMarshaler = &jsonpb.Marshaler{}
)

// PayloadUnaryServerInterceptor returns a new unary server interceptors that logs the payloads of requests.
//
// This *only* works when placed *after* the `grpc_zerolog.UnaryServerInterceptor`. However, the logging can be done to a
// separate instance of the logger.
func PayloadUnaryServerInterceptor(logger *zerolog.Logger, decider grpc_logging.ServerPayloadLoggingDecider) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !decider(ctx, info.FullMethod, info.Server) {
			return handler(ctx, req)
		}
		// Use the provided zerolog.Logger for logging but use the fields from context.
		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return *ctxzerolog.Extract(ctx)
		})

		logProtoMessageAsJSON(logger.With(), req, "grpc.request.content", "server request payload logged as grpc.request.content field")
		resp, err := handler(ctx, req)
		if err == nil {
			logProtoMessageAsJSON(logger.With(), resp, "grpc.response.content", "server response payload logged as grpc.request.content field")
		}
		return resp, err
	}
}

// PayloadStreamServerInterceptor returns a new server server interceptors that logs the payloads of requests.
//
// This *only* works when placed *after* the `grpc_zerolog.StreamServerInterceptor`. However, the logging can be done to a
// separate instance of the logger.
func PayloadStreamServerInterceptor(logger *zerolog.Logger, decider grpc_logging.ServerPayloadLoggingDecider) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !decider(stream.Context(), info.FullMethod, srv) {
			return handler(srv, stream)
		}
		// Use the provided zerolog.Logger for logging but use the fields from context.
		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return *ctxzerolog.Extract(stream.Context())
		})

		newStream := &loggingServerStream{ServerStream: stream, logContext: logger.With()}
		return handler(srv, newStream)
	}
}

// PayloadUnaryClientInterceptor returns a new unary client interceptor that logs the payloads of requests and responses.
func PayloadUnaryClientInterceptor(logger *zerolog.Logger, decider grpc_logging.ClientPayloadLoggingDecider) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !decider(ctx, method) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		logContext := logger.With().Fields(newClientLoggerFields(ctx, method))

		logProtoMessageAsJSON(logContext, req, "grpc.request.content", "client request payload logged as grpc.request.content")
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			logProtoMessageAsJSON(logContext, reply, "grpc.response.content", "client response payload logged as grpc.response.content")
		}
		return err
	}
}

// PayloadStreamClientInterceptor returns a new streaming client interceptor that logs the payloads of requests and responses.
func PayloadStreamClientInterceptor(logger *zerolog.Logger, decider grpc_logging.ClientPayloadLoggingDecider) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if !decider(ctx, method) {
			return streamer(ctx, desc, cc, method, opts...)
		}
		// Use the provided zerolog.Logger for logging but use the fields from context.
		logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
			ctxContext := *ctxzerolog.Extract(ctx)
			ctxContext = ctxContext.Fields(newClientLoggerFields(ctx, method))
			return ctxContext
		})

		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		newStream := &loggingClientStream{ClientStream: clientStream, logContext: logger.With()}
		return newStream, err
	}
}

type loggingClientStream struct {
	grpc.ClientStream
	logContext zerolog.Context
}

func (l *loggingClientStream) SendMsg(m interface{}) error {
	err := l.ClientStream.SendMsg(m)
	if err == nil {
		logProtoMessageAsJSON(l.logContext, m, "grpc.request.content", "server request payload logged as grpc.request.content field")
	}
	return err
}

func (l *loggingClientStream) RecvMsg(m interface{}) error {
	err := l.ClientStream.RecvMsg(m)
	if err == nil {
		logProtoMessageAsJSON(l.logContext, m, "grpc.response.content", "server response payload logged as grpc.response.content field")
	}
	return err
}

type loggingServerStream struct {
	grpc.ServerStream
	logContext zerolog.Context
}

func (l *loggingServerStream) SendMsg(m interface{}) error {
	err := l.ServerStream.SendMsg(m)
	if err == nil {
		logProtoMessageAsJSON(l.logContext, m, "grpc.response.content", "server response payload logged as grpc.response.content field")
	}
	return err
}

func (l *loggingServerStream) RecvMsg(m interface{}) error {
	err := l.ServerStream.RecvMsg(m)
	if err == nil {
		logProtoMessageAsJSON(l.logContext, m, "grpc.request.content", "server request payload logged as grpc.request.content field")
	}
	return err
}

func logProtoMessageAsJSON(logContext zerolog.Context, pbMsg interface{}, key string, msg string) {
	if p, ok := pbMsg.(proto.Message); ok {
		pb := &jsonpbMarshalleble{p}
		blob, err := pb.MarshalJSON()
		if err != nil {
			return
		}
		logger := logContext.Logger()
		logger.Info().Bytes(key, blob).Msg(msg)
	}
}

type jsonpbMarshalleble struct {
	proto.Message
}

func (j *jsonpbMarshalleble) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := JSONPbMarshaller.Marshal(b, j.Message); err != nil {
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}
	return b.Bytes(), nil
}
