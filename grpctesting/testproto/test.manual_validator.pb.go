// Manual code for validation tests.

package grpc_middleware_testproto

import "errors"

func (m *PingRequest) Validate() error {
	if m.SleepTimeMs > 10000 {
		return errors.New("cannot sleep for more than 10s")
	}
	return nil
}
