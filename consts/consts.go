package consts

import "time"

const (
	// MaxSizeOfCache - max size of cache
	MaxSizeOfCache = 1 << 15 // 32k
)

const (
	// DefaultTickerDuration - ticker for cache write file
	DefaultTickerDuration = 1 * time.Second
)
