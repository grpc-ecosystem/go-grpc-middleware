package logging

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tags"
)

// var (
// 	// JsonPbMarshaller is the marshaller used for serializing protobuf messages.
// 	// If needed, this variable can be reassigned with a different marshaller with the same Marshal() signature.
// 	JsonPbMarshaller JsonPbMarshaler = jsonpb.MarshalOptions{}
// )

type serverPayloadReporter struct {
	ctx    context.Context
	logger Logger
}

func (c *serverPayloadReporter) PostCall(error, time.Duration) {}

func (c *serverPayloadReporter) PostMsgSend(req interface{}, err error, duration time.Duration) {
	if err != nil {
		return
	}
	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	// For server send message is the response.
	logProtoMessageAsJson(logger.With("grpc.send.duration", duration.String()), req, "grpc.response.content", "response payload logged as grpc.response.content field")
}

func (c *serverPayloadReporter) PostMsgReceive(reply interface{}, err error, duration time.Duration) {
	if err != nil {
		return
	}
	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	// For server recv message is the request.
	logProtoMessageAsJson(logger.With("grpc.recv.duration", duration.String()), reply, "grpc.request.content", "request payload logged as grpc.request.content field")
}

type clientPayloadReporter struct {
	ctx    context.Context
	logger Logger
}

func (c *clientPayloadReporter) PostCall(error, time.Duration) {}

func (c *clientPayloadReporter) PostMsgSend(req interface{}, err error, duration time.Duration) {
	if err != nil {
		return
	}
	fmt.Printf("PPPPPPP %v",req)
	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	logProtoMessageAsJson(logger.With("grpc.send.duration", duration.String()), req, "grpc.request.content", "request payload logged as grpc.request.content field")
}

func (c *clientPayloadReporter) PostMsgReceive(reply interface{}, err error, duration time.Duration) {
	if err != nil {
		return
	}
	fmt.Printf("YYYYY %v",reply)
	logger := c.logger.With(extractFields(tags.Extract(c.ctx))...)
	logProtoMessageAsJson(logger.With("grpc.recv.duration", duration.String()), reply, "grpc.response.content", "response payload logged as grpc.response.content field")
}

type payloadReportable struct {
	clientDecider ClientPayloadLoggingDecider
	serverDecider ServerPayloadLoggingDecider
	logger        Logger
}

func (r *payloadReportable) ServerReporter(ctx context.Context, req interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	if !r.serverDecider(ctx, interceptors.FullMethod(service, method), req) {
		return interceptors.NoopReporter{}, ctx
	}
	fields := commonFields(KindServerFieldValue, typ, service, method)
	fields = append(fields, "grpc.start_time", time.Now().Format(time.RFC3339))
	if d, ok := ctx.Deadline(); ok {
		fields = append(fields, "grpc.request.deadline", d.Format(time.RFC3339))
	}
	return &serverPayloadReporter{
		ctx:    ctx,
		logger: r.logger.With(fields...),
	}, ctx
}
func (r *payloadReportable) ClientReporter(ctx context.Context, _ interface{}, typ interceptors.GRPCType, service string, method string) (interceptors.Reporter, context.Context) {
	if !r.clientDecider(ctx, interceptors.FullMethod(service, method)) {
		return interceptors.NoopReporter{}, ctx
	}
	fields := commonFields(KindClientFieldValue, typ, service, method)
	fields = append(fields, "grpc.start_time", time.Now().Format(time.RFC3339))
	if d, ok := ctx.Deadline(); ok {
		fields = append(fields, "grpc.request.deadline", d.Format(time.RFC3339))
	}
	return &clientPayloadReporter{
		ctx:    ctx,
		logger: r.logger.With(fields...),
	}, ctx
}

// PayloadUnaryServerInterceptor returns a new unary server interceptors that logs the payloads of requests on INFO level.
// Logger tags will be used from tags context.
func PayloadUnaryServerInterceptor(logger Logger, decider ServerPayloadLoggingDecider) grpc.UnaryServerInterceptor {
	return interceptors.UnaryServerInterceptor(&payloadReportable{logger: logger, serverDecider: decider})
}

// PayloadStreamServerInterceptor returns a new server server interceptors that logs the payloads of requests on INFO level.
// Logger tags will be used from tags context.
func PayloadStreamServerInterceptor(logger Logger, decider ServerPayloadLoggingDecider) grpc.StreamServerInterceptor {
	return interceptors.StreamServerInterceptor(&payloadReportable{logger: logger, serverDecider: decider})
}

// PayloadUnaryClientInterceptor returns a new unary client interceptor that logs the paylods of requests and responses on INFO level.
// Logger tags will be used from tags context.
func PayloadUnaryClientInterceptor(logger Logger, decider ClientPayloadLoggingDecider) grpc.UnaryClientInterceptor {
	return interceptors.UnaryClientInterceptor(&payloadReportable{logger: logger, clientDecider: decider})
}

// PayloadStreamClientInterceptor returns a new streaming client interceptor that logs the paylods of requests and responses on INFO level.
// Logger tags will be used from tags context.
func PayloadStreamClientInterceptor(logger Logger, decider ClientPayloadLoggingDecider) grpc.StreamClientInterceptor {
	return interceptors.StreamClientInterceptor(&payloadReportable{logger: logger, clientDecider: decider})
}

func logProtoMessageAsJson(logger Logger, pbMsg interface{}, key string, msg string) {
	if _, ok := pbMsg.(proto.Message); ok {
		pp := pbMsg.(proto.Message)
		fmt.Printf("VV - %v",pp)
		payload, err := proto.Marshal(pp)
		if err != nil {
			logger = logger.With(key, err.Error())
		} else {
			logger = logger.With(key, string(payload))
		}
		logger.Log(INFO, msg)
	}
}

type jsonpbObjectMarshaler struct {
	pb proto.Message
}

func (j *jsonpbObjectMarshaler) marshalJSON() ([]byte, error) {

	
	b, err := proto.Marshal(j.pb)
	
	if err != nil {
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}
	return b, nil
}
