// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package tracing

type Key string

// KeyValue holds a key and value pair.
type KeyValue struct {
	Key   Key
	Value interface{}
}

// Value creates a KeyValue instance with a Value.
// It supports string, bool, int, int64, float64, string.
func (k Key) Value(v interface{}) KeyValue {
	return KeyValue{
		Key:   k,
		Value: v,
	}
}
