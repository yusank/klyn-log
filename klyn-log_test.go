package klynlog

import (
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	l := &LoggerConfig{
		Prefix:    "KLYN",
		FlushMode: FlushModeEveryLog,
	}
	logger := NewLogger(l)
	// logger := DefaultLogger()

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

func BenchmarkNewLoggerMode1(b *testing.B) {
	l := &LoggerConfig{
		Prefix:    "KLYN",
		FlushMode: FlushModeEveryLog,
	}

	logger := NewLogger(l)
	for i := 0; i < b.N; i++ {
		logger.Error(map[string]interface{}{
			"ip":     "127.0.0.1",
			"userId": i,
		})
	}
}

func BenchmarkNewLoggerMode2(b *testing.B) {
	l := &LoggerConfig{
		Prefix:    "KLYN",
		FlushMode: FlushModeByDuration,
	}

	logger := NewLogger(l)
	for i := 0; i < b.N; i++ {
		logger.Error(map[string]interface{}{
			"ip":     "127.0.0.1",
			"userId": i,
		})
	}
}

func BenchmarkNewLoggerMode3(b *testing.B) {
	l := &LoggerConfig{
		Prefix:    "KLYN",
		FlushMode: FlushModeBySize,
	}

	logger := NewLogger(l)
	for i := 0; i < b.N; i++ {
		logger.Error(map[string]interface{}{
			"ip":     "127.0.0.1",
			"userId": i,
		})
	}
}
