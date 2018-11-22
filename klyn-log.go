package klynlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

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

// KlynLog - implement Logger and provide cache
type KlynLog struct {
	config    *LoggerConfig
	logWriter io.Writer
	cache     *logCache
}

type logCache struct {
	buf       *bytes.Buffer
	cacheLock *sync.RWMutex
}

const (
	// MaxSizeOfCache - max size of cache
	MaxSizeOfCache = 1 << 16 // 64k
)

// LoggerConfig - logger config
type LoggerConfig struct {
	isOff      bool
	isUseCache bool
	Prefix     string
	lock       *sync.RWMutex
}

// NewLogger return Logger
func NewLogger(l *LoggerConfig) Logger {
	l.lock = new(sync.RWMutex)
	return &KlynLog{
		config: l,
	}
}

func (kl *KlynLog) isOff() bool {
	kl.config.lock.RLock()
	defer kl.config.lock.RUnlock()

	return kl.config.isOff
}

func (kl *KlynLog) writeCache(b []byte) {
	kl.cache.cacheLock.Lock()
	kl.cache.buf.Write(b)
	kl.cache.cacheLock.Unlock()
}

func (kl *KlynLog) syncAndFlushCache() error {
	kl.cache.cacheLock.Lock()
	defer kl.cache.cacheLock.Unlock()

	cache := make([]byte, kl.cache.buf.Len())
	_, err := kl.cache.buf.Read(cache)
	if err != nil {
		return err
	}

	_, err = kl.logWriter.Write(cache)
	if err != nil {
		return err
	}

	kl.cache.buf.Reset()
	return nil
}

func (kl *KlynLog) cacheLen() int {
	kl.cache.cacheLock.RLock()
	defer kl.cache.cacheLock.RUnlock()

	return kl.cache.buf.Len()
}

func (kl *KlynLog) setLogOff() {
	kl.config.lock.Lock()
	kl.config.isOff = true
	kl.config.lock.RLock()
}

// Trace - trace level log
func (kl *KlynLog) Trace(j interface{}) {
	kl.log(LoggerLevelTrace, j)
}

// Debug - debug level log
func (kl *KlynLog) Debug(j interface{}) {
	kl.log(LoggerLevelDebug, j)
}

// Info -
func (kl *KlynLog) Info(j interface{}) {
	kl.log(LoggerLevelInfo, j)
}

// Warn -
func (kl *KlynLog) Warn(j interface{}) {
	kl.log(LoggerLevelWarn, j)
}

// Error -
func (kl *KlynLog) Error(j interface{}) {
	kl.log(LoggerLevelError, j)
}

// Fatal -
func (kl *KlynLog) Fatal(j interface{}) {
	kl.log(LoggerLevelFatal, j)
}

// Any -
func (kl *KlynLog) Any(level int, j interface{}) {
	kl.log(level, j)
}

// OFF - off all level log
func (kl *KlynLog) OFF() {
	kl.setLogOff()
}

func (kl *KlynLog) log(level int, j interface{}) {
	if kl.isOff() {
		return
	}

	b, _ := json.Marshal(j)

	line := fmt.Sprintf("[%s] | LEVEL:%d | message:%s\n", kl.config.Prefix, level, string(b))
	file, err := kl.getLogFile()
	if err != nil {
		panic(err)
	}

	file.Write([]byte(line))

	return
}

func (kl *KlynLog) getLogFile() (file *os.File, err error) {
	return os.Create(fmt.Sprintf("%s.log", kl.config.Prefix))
}

func (kl *KlynLog) monitor() {
	for {
		if kl.cacheLen() >= MaxSizeOfCache {
			if err := kl.syncAndFlushCache(); err != nil {
				panic(err)
			}
		}
	}
}
