// Manual code for logging field extraction tests.

package mwitkow_testproto

// This is implementing grpc_logging.requestLogFieldsExtractor
func (m *PingRequest) ExtractRequestFields() map[string]interface{} {
	return map[string]interface{}{"value": m.Value}
}
