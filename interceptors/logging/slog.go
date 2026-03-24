// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Copied from https://cs.opensource.google/go/x/exp/+/10a50721:slog/level.go

package logging

import "log/slog"

// Level is an alias for slog.Level.
type Level slog.Level

// Common level values.
const (
	LevelDebug Level = Level(slog.LevelDebug)
	LevelInfo  Level = Level(slog.LevelInfo)
	LevelWarn  Level = Level(slog.LevelWarn)
	LevelError Level = Level(slog.LevelError)
)
