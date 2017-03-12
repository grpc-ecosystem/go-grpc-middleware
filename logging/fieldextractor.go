// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_logging

import (
	"reflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// ErrorToCode function determines the error code of an error
// This makes using custom errors with grpc middleware easier
type ErrorToCode func(err error) codes.Code

func DefaultErrorToCode(err error) codes.Code {
	return grpc.Code(err)
}

// RequestLogFieldExtractorFunc is a user-provided function that extracts field information from a gRPC request.
// It is called from every logging middleware on arrival of unary request or a server-stream request.
// Keys and values will be added to the logging request context.
type RequestLogFieldExtractorFunc func(fullMethod string, req interface{}) (keys []string, values []interface{})

type requestLogFieldsExtractor interface {
	// ExtractLogFields is a method declared on a Protobuf message that extracts log fields from the interface.
	ExtractLogFields() (keys []string, values []interface{})
}

// CodeGenRequestLogFieldExtractor is a function that relies on code-generated functions that export log fields from requests.
// These are usually coming from a protoc-plugin that generates additional information based on custom field options.
func CodeGenRequestLogFieldExtractor(fullMethod string, req interface{}) (keys []string, values []interface{}) {
	if ext, ok := req.(requestLogFieldsExtractor); ok {
		return ext.ExtractLogFields()
	}
	return nil, nil
}

// TagedRequestFiledExtractor is a function that relies on Go struct tags to export log fields from requests.
// These are usualy coming from a protoc-plugin, such as Gogo protobuf.
//
//  message Metadata {
//     repeated string tags = 1 [ (gogoproto.moretags) = "log_field:\"meta_tags\"" ];
//  }
//
// It requires the tag to be `log_field` and is recursively executed through all non-repeated structs.
func TagedRequestFiledExtractor(fullMethod string, req interface{}) (keys []string, values []interface{}) {
	return reflectMessageTags(req)
}

func reflectMessageTags(msg interface{}) (keys []string, values []interface{}) {
	v := reflect.ValueOf(msg)
	// Only deal with pointers to structs.
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, nil
	}
	// Deref the pointer get to the struct.
	v = v.Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		kind := field.Kind()
		// Only recurse down direct pointers, which should only be to nested structs.
		if kind == reflect.Ptr {
			k, v := reflectMessageTags(field.Interface())
			keys = append(keys, k...)
			values = append(values, v...)
		}
		// In case of arrays/splices (repeated fields) go down to the concrete type.
		if kind == reflect.Array || kind == reflect.Slice {
			if field.Len() == 0 {
				continue
			}
			kind = field.Index(0).Kind()
		}
		// Only be interested in
		if (kind >= reflect.Bool && kind <= reflect.Float64) || kind == reflect.String {
			if tag := t.Field(i).Tag.Get("log_field"); tag != "" {
				keys = append(keys, tag)
				values = append(values, field.Interface())
			}
		}
	}
	return
}
