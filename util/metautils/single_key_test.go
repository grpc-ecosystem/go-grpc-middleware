package metautils_test

import (
	"testing"

	"github.com/mwitkow/go-grpc-middleware/util/metautils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

var (
	parentCtx = context.WithValue(context.TODO(), "parentKey", "parentValue")
)

func assertRetainsParentContext(t *testing.T, ctx context.Context) {
	x := ctx.Value("parentKey")
	assert.EqualValues(t, "parentValue", x, "context must contain parentCtx")
}

func TestSingleReadYourWrites(t *testing.T) {
	key := "someKey"
	value := "123456"
	c := metautils.SetSingle(parentCtx, key, value)
	assertRetainsParentContext(t, c)
	out, ok := metautils.GetSingle(c, key)
	assert.True(t, ok, "GetSingle should find the key")
	assert.Equal(t, value, out, "value from GetSingle must match SetSingle")
}

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
