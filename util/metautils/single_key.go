// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package metautils

import (
	"encoding/base64"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

const (
	binHdrSuffix = "-bin"
)

// GetSingle extracts a single-value metadata key from Context.
// First return is the value of the key, followed by a bool indicator.
// The bool indicator being false means the string should be discarded. It can be false if
// the context has no metadata at all, the key in metadata doesn't exist or there are multiple values.
// Deprecated, use NiceMD.Get.
func GetSingle(ctx context.Context, keyName string) (string, bool) {
	// TODO(mwitkow): Fix binary content support.
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return "", false
	}
	keyName = strings.ToLower(keyName)
	valSlice, ok := md[keyName]
	if !ok {
		return "", false
	}
	if len(valSlice) != 1 {
		return "", false
	}
	return valSlice[0], true
}

// SetSingle sets or overrides a metadata key to be single value in the Context.
// It returns a new context.Context object that contains a *copy* of the metadata inside the given
// context.
// Deprecated, use NiceMD.Set.
func SetSingle(ctx context.Context, keyName string, keyValue string) context.Context {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = metadata.Pairs(keyName, keyValue)
		return metadata.NewContext(ctx, md)
	}
	k, v := encodeKeyValue(keyName, keyValue)
	md[k] = []string{v}
	return ctx // we use the same context because we modified the metadata in place.
}

func encodeKeyValue(k, v string) (string, string) {
	k = strings.ToLower(k)
	if strings.HasSuffix(k, binHdrSuffix) {
		val := base64.StdEncoding.EncodeToString([]byte(v))
		v = string(val)
	}
	return k, v
}
