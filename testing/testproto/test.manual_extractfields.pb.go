// Manual code for logging field extraction tests.

package mwitkow_testproto

// This is implementing grpc_logging.requestLogFieldsExtractor
func (m *PingRequest) ExtractLogFields() (keys []string, values []interface{}) {
	return []string{"request.value"}, []interface{}{m.Value}
}
