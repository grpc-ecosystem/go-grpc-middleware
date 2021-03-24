package ctxzap

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestExtractS(t *testing.T) {
	ctx := context.Background()
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	decoratedCtx := ToContext(ctx, logger)

	sl := ExtractS(decoratedCtx)
	assert.NotNil(t, t, sl)
}
