package klynlog

import (
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	logger := DefaultLogger()

	go func() {
		for i := 0; i < 1000; i++ {
			logger.Warn(map[string]interface{}{
				"name":   "hello world",
				"userId": i,
				"event": map[string]interface{}{
					"gameId": "dddjs",
				},
			})
		}
	}()

	for i := 0; i < 1000; i++ {
		logger.Error(map[string]interface{}{
			"ip":     "127.0.0.1",
			"userId": i,
		})
	}

	time.Sleep(2 * time.Second)
}

func TestLogFLush(t *testing.T) {
	conf := &LoggerConfig{
		FlushMode: FlushModeEveryLog,
		IsDebug:   true,
		Prefix:    "yusank",
	}

	logger := NewLogger(conf)
	for i := 0; i < 1000; i++ {
		logger.Error(map[string]interface{}{
			"ip":     "127.0.0.1",
			"userId": i,
		})

		time.Sleep(time.Millisecond * 10)
	}
}
