//go:build !retrynotrace

// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package retry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/trace"
)

func Test_traceFromCtx(t *testing.T) {
	tr := trace.New("test", "with trace")
	ctx := trace.NewContext(context.Background(), tr)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name  string
		args  args
		want  trace.Trace
		want1 bool
	}{
		{
			name:  "should return trace",
			args:  args{ctx: ctx},
			want:  tr,
			want1: true,
		},
		{
			name:  "should return false if trace not found in ctx",
			args:  args{ctx: context.Background()},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := traceFromCtx(tt.args.ctx)
			assert.Equalf(t, tt.want, got, "traceFromCtx(%v)", tt.args.ctx)
			assert.Equalf(t, tt.want1, got1, "traceFromCtx(%v)", tt.args.ctx)
		})
	}
}
