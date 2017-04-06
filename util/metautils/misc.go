// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package metautils

import "google.golang.org/grpc/metadata"

// Copy creates a shallow copy of the metadata.
func Copy(parent metadata.MD) metadata.MD {
	if parent == nil {
		return nil
	}
	ret := metadata.New(nil)
	for k, vv := range parent {
		ret[k] = vv
	}
	return ret
}
