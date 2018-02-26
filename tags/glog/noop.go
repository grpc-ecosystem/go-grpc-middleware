// Copyright 2018 AppsCode Inc. All Rights Reserved.
// See LICENSE for licensing terms.

package ctx_glog

import (
	"io/ioutil"

	"google.golang.org/grpc/grpclog"
)

var (
	nullLogger = grpclog.NewLoggerV2(ioutil.Discard, ioutil.Discard, ioutil.Discard)
)
