package grpc_logrus

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestDurationToTimeMillisField(t *testing.T) {
	_, val := DurationToTimeMillisField(time.Microsecond * 100)
	assert.Equal(t, val.(float32), float32(0.1), "sub millisecond values should be correct")
}

func TestDefaultMessageProducer(t *testing.T) {
	t.Parallel()

	errNotFound := errors.New("not found")
	testcases := []struct {
		label  string
		format string
		level  logrus.Level
		code   codes.Code
		err    error
		fields logrus.Fields
	}{
		{
			label:  "should emit info message without error",
			format: "test",
			level:  logrus.InfoLevel,
			code:   codes.OK,
			fields: logrus.Fields{
				"foo": "bar",
			},
		},
		{
			label:  "should emit info message with error",
			format: "test",
			level:  logrus.InfoLevel,
			code:   codes.NotFound,
			err:    errNotFound,
			fields: logrus.Fields{
				"foo":           "bar",
				logrus.ErrorKey: errNotFound,
			},
		},
		{
			label:  "should emit trace message without error",
			format: "test",
			level:  logrus.TraceLevel,
			code:   codes.OK,
			fields: logrus.Fields{
				"foo": "bar",
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			logger, hook := test.NewNullLogger()

			logger.SetLevel(logrus.TraceLevel)

			logrusEntry := logger.WithField("foo", "bar")

			ctx := ctxlogrus.ToContext(context.Background(), logrusEntry)

			DefaultMessageProducer(ctx, tc.format, tc.level, tc.code, tc.err, logrus.Fields{})

			lastEntry := hook.LastEntry()

			require.NotNil(t, lastEntry)

			assert.Equal(t, "test", lastEntry.Message)
			assert.Equal(t, tc.fields, lastEntry.Data)
			assert.Equal(t, tc.level, lastEntry.Level)
		})
	}
}
