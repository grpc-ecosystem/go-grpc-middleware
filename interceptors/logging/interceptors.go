// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logging

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/proto"
)

type reporter struct {
	interceptors.CallMeta

	ctx             context.Context
	kind            string
	startCallLogged bool

	opts   *options
	fields Fields
	logger Logger
}

func (c *reporter) PostCall(err error, duration time.Duration) {
	if !has(c.opts.loggableEvents, FinishCall) {
		return
	}
	if err == io.EOF {
		err = nil
	}

	code := c.opts.codeFunc(err)
	fields := c.fields.WithUnique(ExtractFields(c.ctx))
	fields = fields.AppendUnique(Fields{"grpc.code", code.String()})
	if err != nil {
		fields = fields.AppendUnique(Fields{"grpc.error", fmt.Sprintf("%v", err)})
	}
	c.logger.Log(c.ctx, c.opts.levelFunc(code), "finished call", fields.AppendUnique(c.opts.durationFieldFunc(duration))...)
}

func (c *reporter) PostMsgSend(payload any, err error, duration time.Duration) {
	logLvl := c.opts.levelFunc(c.opts.codeFunc(err))
	fields := c.fields.WithUnique(ExtractFields(c.ctx))
	if err != nil {
		fields = fields.AppendUnique(Fields{"grpc.error", fmt.Sprintf("%v", err)})
	}
	if !c.startCallLogged && has(c.opts.loggableEvents, StartCall) {
		c.startCallLogged = true
		c.logger.Log(c.ctx, logLvl, "started call", fields.AppendUnique(c.opts.durationFieldFunc(duration))...)
	}

	if err != nil || !has(c.opts.loggableEvents, PayloadSent) {
		return
	}
	if c.CallMeta.IsClient {
		p, ok := payload.(proto.Message)
		if !ok {
			c.logger.Log(
				c.ctx,
				LevelError,
				"payload is not a google.golang.org/protobuf/proto.Message; programmatic error?",
				fields.AppendUnique(Fields{"grpc.request.type", fmt.Sprintf("%T", payload)}),
			)
			return
		}

		fields = fields.AppendUnique(Fields{"grpc.send.duration", duration.String(), "grpc.request.content", p})
		c.logger.Log(c.ctx, logLvl, "request sent", fields...)
	} else {
		p, ok := payload.(proto.Message)
		if !ok {
			c.logger.Log(
				c.ctx,
				LevelError,
				"payload is not a google.golang.org/protobuf/proto.Message; programmatic error?",
				fields.AppendUnique(Fields{"grpc.response.type", fmt.Sprintf("%T", payload)}),
			)
			return
		}

		fields = fields.AppendUnique(Fields{"grpc.send.duration", duration.String(), "grpc.response.content", p})
		c.logger.Log(c.ctx, logLvl, "response sent", fields...)
	}
}

func (c *reporter) PostMsgReceive(payload any, err error, duration time.Duration) {
	logLvl := c.opts.levelFunc(c.opts.codeFunc(err))
	fields := c.fields.WithUnique(ExtractFields(c.ctx))
	if err != nil {
		fields = fields.AppendUnique(Fields{"grpc.error", fmt.Sprintf("%v", err)})
	}
	if !c.startCallLogged && has(c.opts.loggableEvents, StartCall) {
		c.startCallLogged = true
		c.logger.Log(c.ctx, logLvl, "started call", fields.AppendUnique(c.opts.durationFieldFunc(duration))...)
	}

	if err != nil || !has(c.opts.loggableEvents, PayloadReceived) {
		return
	}
	if !c.CallMeta.IsClient {
		p, ok := payload.(proto.Message)
		if !ok {
			c.logger.Log(
				c.ctx,
				LevelError,
				"payload is not a google.golang.org/protobuf/proto.Message; programmatic error?",
				fields.AppendUnique(Fields{"grpc.request.type", fmt.Sprintf("%T", payload)}),
			)
			return
		}

		fields = fields.AppendUnique(Fields{"grpc.recv.duration", duration.String(), "grpc.request.content", p})
		c.logger.Log(c.ctx, logLvl, "request received", fields...)
	} else {
		p, ok := payload.(proto.Message)
		if !ok {
			c.logger.Log(
				c.ctx,
				LevelError,
				"payload is not a google.golang.org/protobuf/proto.Message; programmatic error?",
				fields.AppendUnique(Fields{"grpc.response.type", fmt.Sprintf("%T", payload)}),
			)
			return
		}

		fields = fields.AppendUnique(Fields{"grpc.recv.duration", duration.String(), "grpc.response.content", p})
		c.logger.Log(c.ctx, logLvl, "response received", fields...)
	}
}

func reportable(logger Logger, opts *options) interceptors.CommonReportableFunc {
	return func(ctx context.Context, c interceptors.CallMeta) (interceptors.Reporter, context.Context) {
		kind := KindServerFieldValue
		if c.IsClient {
			kind = KindClientFieldValue
		}

		fields := ExtractFields(ctx).WithUnique(newCommonFields(kind, c))
		if !c.IsClient {
			if peer, ok := peer.FromContext(ctx); ok {
				fields = append(fields, "peer.address", peer.Addr.String())
			}
		}
		if opts.fieldsFromCtxFn != nil {
			fields = fields.AppendUnique(opts.fieldsFromCtxFn(ctx))
		}

		singleUseFields := Fields{"grpc.start_time", time.Now().Format(opts.timestampFormat)}
		if d, ok := ctx.Deadline(); ok {
			singleUseFields = singleUseFields.AppendUnique(Fields{"grpc.request.deadline", d.Format(opts.timestampFormat)})
		}
		return &reporter{
			CallMeta:        c,
			ctx:             ctx,
			startCallLogged: false,
			opts:            opts,
			fields:          fields.WithUnique(singleUseFields),
			logger:          logger,
			kind:            kind,
		}, InjectFields(ctx, fields)
	}
}

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
// Logger will read existing and write new logging.Fields available in current context.
// See `ExtractFields` and `InjectFields` for details.
func UnaryClientInterceptor(logger Logger, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return interceptors.UnaryClientInterceptor(reportable(logger, o))
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
// Logger will read existing and write new logging.Fields available in current context.
// See `ExtractFields` and `InjectFields` for details.
func StreamClientInterceptor(logger Logger, opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateClientOpt(opts)
	return interceptors.StreamClientInterceptor(reportable(logger, o))
}

// UnaryServerInterceptor returns a new unary server interceptors that optionally logs endpoint handling.
// Logger will read existing and write new logging.Fields available in current context.
// See `ExtractFields` and `InjectFields` for details.
func UnaryServerInterceptor(logger Logger, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return interceptors.UnaryServerInterceptor(reportable(logger, o))
}

// StreamServerInterceptor returns a new stream server interceptors that optionally logs endpoint handling.
// Logger will read existing and write new logging.Fields available in current context.
// See `ExtractFields` and `InjectFields` for details..
func StreamServerInterceptor(logger Logger, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return interceptors.StreamServerInterceptor(reportable(logger, o))
}
