// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package tracing

import (
	"context"
	"io"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type reporter struct {
	ctx  context.Context
	span Span

	receivedMessageID int
	sentMessageID     int
}

func (o *reporter) PostCall(err error, _ time.Duration) {
	// Finish span.
	if err != nil && err != io.EOF {
		s, _ := status.FromError(err)
		o.span.SetStatus(s.Code(), s.Message())
		o.span.SetAttributes(statusCodeAttr(s.Code()))
	} else {
		o.span.SetAttributes(statusCodeAttr(codes.OK))
	}
	o.span.End()
}

func (o *reporter) PostMsgSend(payload interface{}, err error, d time.Duration) {
	o.sentMessageID++
	addEvent(o.span, RPCMessageTypeSent, o.sentMessageID, payload)
}

func (o *reporter) PostMsgReceive(payload interface{}, err error, d time.Duration) {
	o.receivedMessageID++
	addEvent(o.span, RPCMessageTypeReceived, o.receivedMessageID, payload)
}

func addEvent(span Span, messageType KeyValue, messageID int, payload interface{}) {
	if p, ok := payload.(proto.Message); ok {
		span.AddEvent("message",
			messageType,
			rpcMessageIDKey.Value(messageID),
			rpcMessageUncompressedSizeKey.Value(proto.Size(p)),
		)
		return
	}
	span.AddEvent("message",
		messageType,
		rpcMessageIDKey.Value(messageID),
	)
}

// statusCodeAttr returns status code attribute based on given gRPC code
func statusCodeAttr(c codes.Code) KeyValue {
	return grpcStatusCodeKey.Value(int64(c))
}
