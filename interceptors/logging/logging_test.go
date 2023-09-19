// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logging

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFieldsInjectExtractFromContext(t *testing.T) {
	c := context.Background()
	f := ExtractFields(c)
	require.Equal(t, Fields(nil), f)

	f = f.AppendUnique([]any{"a", "1", "b", "2"})
	require.Equal(t, Fields{"a", "1", "b", "2"}, f)

	c2 := InjectFields(c, f)

	// First context should be untouched.
	f = ExtractFields(c)
	require.Equal(t, Fields(nil), f)
	f = ExtractFields(c2)
	require.Equal(t, Fields{"a", "1", "b", "2"}, f)

	f = Fields{"a", "changed"}.WithUnique(f)
	require.Equal(t, Fields{"a", "changed", "b", "2"}, f)

	c3 := InjectFields(c, f)

	// Old contexts should be untouched.
	f = ExtractFields(c)
	require.Equal(t, Fields(nil), f)
	f = ExtractFields(c2)
	require.Equal(t, Fields{"a", "1", "b", "2"}, f)
	f = ExtractFields(c3)
	require.Equal(t, Fields{"a", "changed", "b", "2"}, f)
}

func TestFieldsDelete(t *testing.T) {
	f := Fields{"a", "1", "b", "2"}
	f.Delete("a")
	require.Equal(t, Fields{"b", "2"}, f)
	f.Delete("b")
	require.Equal(t, Fields{}, f)
	f.Delete("c")
	require.Equal(t, Fields{}, f)
}
