// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package consts

import "time"

const (
	// MaxSizeOfCache - max size of cache
	MaxSizeOfCache = 1 << 15 // 32k
)

const (
	// DefaultTickerDuration - ticker for cache write file
	DefaultTickerDuration = 200 * time.Millisecond
)

const (
	// DefaultLogDir -
	DefaultLogDir = "logFiles"
)
