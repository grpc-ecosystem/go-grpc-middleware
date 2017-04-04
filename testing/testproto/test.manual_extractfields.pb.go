// Manual code for logging field extraction tests.

package mwitkow_testproto

// This is implementing grpc_logging.requestLogFieldsExtractor
func (m *PingRequest) ExtractRequestFields() (keys []string, values []interface{}) {
	return []string{"value"}, []interface{}{m.Value}
}
