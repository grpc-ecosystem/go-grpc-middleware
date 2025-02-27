// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

//go:build retrynotrace

package retry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_traceFromCtx(t *testing.T) {
	tr := notrace{}
	ctx := context.Background()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name  string
		args  args
		want  notrace
		want1 bool
	}{
		{
			name:  "should return notrace",
			args:  args{ctx: ctx},
			want:  tr,
			want1: true,
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
