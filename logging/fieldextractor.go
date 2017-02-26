// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_logging

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
