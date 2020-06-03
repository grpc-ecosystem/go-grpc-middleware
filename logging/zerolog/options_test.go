package zerolog_test

import (
	"testing"
	"time"

	grpc_zerolog "github.com/irridia/go-grpc-middleware/logging/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestDurationToTimeMillisField(t *testing.T) {
	_, val := grpc_zerolog.DurationToTimeMillisField(time.Microsecond * 100)
	assert.Equal(t, val.(float32), float32(0.1), "sub millisecond values should be correct")
}
