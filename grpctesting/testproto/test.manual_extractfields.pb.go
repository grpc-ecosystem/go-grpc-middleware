// Manual code for logging field extraction tests.

package grpc_middleware_testproto

const TestServiceFullName = "grpc_middleware.testproto.TestService"

// This is implementing ctxtags.requestFieldsExtractor
func (m *PingRequest) ExtractRequestFields(appendToMap map[string]string) {
	appendToMap["value"] = m.Value
}
