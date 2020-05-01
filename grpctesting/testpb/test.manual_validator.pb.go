// Manual code for validation tests.

package testpb

import "errors"

func (m *PingRequest) Validate() error {
	if m.SleepTimeMs > 10000 {
		return errors.New("cannot sleep for more than 10s")
	}
	return nil
}
