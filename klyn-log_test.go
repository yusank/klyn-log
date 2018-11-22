package klynlog

import (
	"testing"
)

func TestLog(t *testing.T) {
	c := &LoggerConfig{
		Prefix: "KLYN",
	}
	logger := NewLogger(c)

	logger.Info(map[string]interface{}{
		"ip":     "127.0.0.1",
		"userId": 123,
	})
}
