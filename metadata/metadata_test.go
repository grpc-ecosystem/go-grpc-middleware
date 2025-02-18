// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package metadata_test

import (
	"context"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/metadata"
	"github.com/stretchr/testify/assert"
	grpcMetadata "google.golang.org/grpc/metadata"
)

type parentKey struct{}

var (
	testPairs = []string{"singlekey", "uno", "multikey", "one", "multikey", "two", "multikey", "three"}
	parentCtx = context.WithValue(context.TODO(), parentKey{}, "parentValue")
)

func assertRetainsParentContext(t *testing.T, ctx context.Context) {
	x := ctx.Value(parentKey{})
	assert.EqualValues(t, "parentValue", x, "context must contain parentCtx")
}

func TestNiceMD_Get(t *testing.T) {
	md := metadata.MD(grpcMetadata.Pairs(testPairs...))
	assert.Equal(t, "uno", md.Get("singlekey"), "for present single-key value it should return it")
	assert.Equal(t, "one", md.Get("multikey"), "for present multi-key should return first value")
	assert.Empty(t, md.Get("nokey"), "for non existing key should return stuff")
}

func TestNiceMD_Del(t *testing.T) {
	md := metadata.MD(grpcMetadata.Pairs(testPairs...))
	assert.Equal(t, "uno", md.Get("singlekey"), "for present single-key value it should return it")
	md.Del("singlekey").Del("doesnt exist")
	assert.Empty(t, md.Get("singlekey"), "after deletion singlekey shouldn't exist")
}

func TestNiceMD_Add(t *testing.T) {
	md := metadata.MD(grpcMetadata.Pairs(testPairs...))
	md.Add("multikey", "four").Add("newkey", "something")
	assert.EqualValues(t, []string{"one", "two", "three", "four"}, md["multikey"], "append should add a new four at the end")
	assert.EqualValues(t, []string{"something"}, md["newkey"], "append should be able to create new keys")
}

func TestNiceMD_Set(t *testing.T) {
	md := metadata.MD(grpcMetadata.Pairs(testPairs...))
	md.Set("multikey", "one").Set("newkey", "something").Set("newkey", "another")
	assert.EqualValues(t, []string{"one"}, md["multikey"], "set should override existing multi keys")
	assert.EqualValues(t, []string{"another"}, md["newkey"], "set should override new keys")
}

func TestNiceMD_Clone(t *testing.T) {
	md := metadata.MD(grpcMetadata.Pairs(testPairs...))
	fullCopied := md.Clone()
	assert.Equal(t, len(fullCopied), len(md), "clone full should copy all keys")
	assert.Equal(t, "uno", fullCopied.Get("singlekey"), "full copied should have content")
	subCopied := md.Clone("multikey")
	assert.Len(t, subCopied, 1, "sub copied clone should only have one key")
	assert.Empty(t, subCopied.Get("singlekey"), "there shouldn't be a singlekey in the subcopied")

	// Test side effects and full copying:
	assert.EqualValues(t, subCopied["multikey"], md["multikey"], "before overwrites multikey should have the same values")
	subCopied["multikey"][1] = "modifiedtwo"
	assert.NotEqual(t, subCopied["multikey"], md["multikey"], "before overwrites multikey should have the same values")
}

func TestNiceMD_ToOutgoing(t *testing.T) {
	md := metadata.MD(grpcMetadata.Pairs(testPairs...))
	nCtx := md.ToOutgoing(parentCtx)
	assertRetainsParentContext(t, nCtx)

	eCtx := metadata.ExtractOutgoing(nCtx).Clone().Set("newvalue", "something").ToOutgoing(nCtx)
	assertRetainsParentContext(t, eCtx)
	assert.NotEqual(t, metadata.ExtractOutgoing(nCtx), metadata.ExtractOutgoing(eCtx), "the niceMD pointed to by ectx and nctx are different.")
}

func TestNiceMD_ToIncoming(t *testing.T) {
	md := metadata.MD(grpcMetadata.Pairs(testPairs...))
	nCtx := md.ToIncoming(parentCtx)
	assertRetainsParentContext(t, nCtx)

	eCtx := metadata.ExtractIncoming(nCtx).Clone().Set("newvalue", "something").ToIncoming(nCtx)
	assertRetainsParentContext(t, eCtx)
	assert.NotEqual(t, metadata.ExtractIncoming(nCtx), metadata.ExtractIncoming(eCtx), "the niceMD pointed to by ectx and nctx are different.")
}
