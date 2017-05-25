package metautils_test

import (
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

func TestSingleFailsReading(t *testing.T) {
	key := "someKey"
	for _, tcase := range []struct {
		caseName string
		ctx      context.Context
	}{
		{
			caseName: "ignores multivalues",
			ctx:      metadata.NewContext(parentCtx, metadata.Pairs(key, "value1", key, "value2")),
		},
		{
			caseName: "handles not found with values",
			ctx:      metadata.NewContext(parentCtx, metadata.Pairs("another value", "value1")),
		},
		{
			caseName: "handles non-MD ctx",
			ctx:      parentCtx,
		},
	} {
		t.Run(tcase.caseName, func(t *testing.T) {
			val, out := metautils.GetSingle(tcase.ctx, key)
			assert.False(t, out, "must return not found")
			assert.Empty(t, val, "the output must be empty")
		})
	}
}

func TestSingleImmutableMD(t *testing.T) {
	key := "someKey"
	value := "123456"
	md := metadata.Pairs(key, value)
	parent := metadata.NewContext(context.Background(), md)
	ctx := metautils.SetSingle(parent, "otherKey", "654321")
	newMD, _ := metadata.FromContext(ctx)
	assert.NotEqual(t, newMD, md, "new MD must be a copy of parent MD")
}
