// Manual code for logging field extraction tests.

package testpb

const TestServiceFullName = "testing.testpb.v1.TestService"

// This is implementing tags.requestFieldsExtractor
func (x *PingRequest) ExtractRequestFields(appendToMap map[string]string) {
	appendToMap["value"] = x.Value
}

// This is implementing tags.requestFieldsExtractor
func (x *PingErrorRequest) ExtractRequestFields(appendToMap map[string]string) {
	appendToMap["value"] = x.Value
}

// This is implementing tags.requestFieldsExtractor
func (x *PingListRequest) ExtractRequestFields(appendToMap map[string]string) {
	appendToMap["value"] = x.Value
}

// This is implementing tags.requestFieldsExtractor
func (x *PingStreamRequest) ExtractRequestFields(appendToMap map[string]string) {
	appendToMap["value"] = x.Value
}
