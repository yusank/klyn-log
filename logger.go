// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package klynlog

import "fmt"

// Logger provide leveled log
type Logger interface {
	Trace(j interface{})
	Debug(j interface{})
	Info(j interface{})
	Warn(j interface{})
	Error(j interface{})
	Fatal(j interface{})
	Any(l Level, j interface{})
	OFF()
}

type LogFunc func(j interface{})

// Level - log level
type Level uint8

const (
	// LoggerLevelTrace - log trace level
	LoggerLevelTrace Level = iota + 1
	// LoggerLevelDebug - log debug level
	LoggerLevelDebug
	// LoggerLevelInfo - log info level
	LoggerLevelInfo
	// LoggerLevelWarn - log warn level
	LoggerLevelWarn
	// LoggerLevelError - log error level
	LoggerLevelError
	// LoggerLevelFatal - log fatal level
	LoggerLevelFatal
)

// String - return level name
func (l Level) String() string {
	switch l {
	case LoggerLevelTrace:
		return "trace"
	case LoggerLevelDebug:
		return "debug"
	case LoggerLevelInfo:
		return "info"
	case LoggerLevelWarn:
		return "warn"
	case LoggerLevelError:
		return "error"
	case LoggerLevelFatal:
		return "fatal"
	default:
		return fmt.Sprint(uint8(l))
	}
}
