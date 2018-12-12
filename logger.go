// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package klynlog

// Logger provide leveled log
type Logger interface {
	Trace(j interface{})
	Debug(j interface{})
	Info(j interface{})
	Warn(j interface{})
	Error(j interface{})
	Fatal(j interface{})
	Any(level int, j interface{})
	OFF()
}

const (
	// LoggerLevelTrace - log trace level
	LoggerLevelTrace = iota + 1
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
