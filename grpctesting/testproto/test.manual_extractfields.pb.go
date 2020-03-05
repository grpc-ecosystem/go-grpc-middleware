// Manual code for logging field extraction tests.

package testproto

const TestServiceFullName = "grpc_middleware.testproto.TestService"

// This is implementing tags.requestFieldsExtractor
func (m *PingRequest) ExtractRequestFields(appendToMap map[string]string) {
	appendToMap["value"] = m.Value
}
