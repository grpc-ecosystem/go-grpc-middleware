// Manual code for validation tests.

package testpb

import "github.com/pkg/errors"

func (x *PingRequest) Validate() error {
	if x.SleepTimeMs > 10000 {
		return errors.New("cannot sleep for more than 10s")
	}
	return nil
}

func (x *PingErrorRequest) Validate() error {
	if x.SleepTimeMs > 10000 {
		return errors.New("cannot sleep for more than 10s")
	}
	return nil
}

func (x *PingListRequest) Validate() error {
	if x.SleepTimeMs > 10000 {
		return errors.New("cannot sleep for more than 10s")
	}
	return nil
}

func (x *PingStreamRequest) Validate() error {
	if x.SleepTimeMs > 10000 {
		return errors.New("cannot sleep for more than 10s")
	}
	return nil
}

var (
	GoodPing       = &PingRequest{Value: "something", SleepTimeMs: 9999}
	GoodPingError  = &PingErrorRequest{Value: "something", SleepTimeMs: 9999}
	GoodPingList   = &PingListRequest{Value: "something", SleepTimeMs: 9999}
	GoodPingStream = &PingStreamRequest{Value: "something", SleepTimeMs: 9999}

	BadPing       = &PingRequest{Value: "something", SleepTimeMs: 10001}
	BadPingError  = &PingErrorRequest{Value: "something", SleepTimeMs: 10001}
	BadPingList   = &PingListRequest{Value: "something", SleepTimeMs: 10001}
	BadPingStream = &PingStreamRequest{Value: "something", SleepTimeMs: 10001}
)
