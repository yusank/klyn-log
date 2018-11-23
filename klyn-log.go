package klynlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"git.yusank.space/yusank/klyn-log/consts"
	"git.yusank.space/yusank/klyn-log/utils"
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
	logWriter *logWriter // log final destination
	cache     *logCache  // log temp cache
}

// LoggerConfig - logger config
type LoggerConfig struct {
	isOff   bool
	IsDebug bool
	Prefix  string
}

type logCache struct {
	buf       *bytes.Buffer
	ticker    *time.Ticker
	errChan   chan error
	cacheLock *sync.RWMutex
}

type logWriter struct {
	writer     *os.File
	writerLock *sync.RWMutex
}

// Closer close log writer if it necessary
type Closer func() error

// NewLogger return Logger
func NewLogger(l *LoggerConfig) Logger {
	cache := &logCache{
		buf:       new(bytes.Buffer),
		cacheLock: new(sync.RWMutex),
		ticker:    time.NewTicker(consts.DefaultTickerDuration),
		errChan:   make(chan error, 0),
	}
	logger := &KlynLog{
		config: l,
		cache:  cache,
	}

	if err := utils.CreateIfNotExist(consts.DefaultLogDir); err != nil {
		panic(err)
	}

	go logger.monitor()

	return logger
}

// DefaultLogger - get default logger
func DefaultLogger() Logger {
	conf := &LoggerConfig{
		Prefix: "KLYN",
	}

	return NewLogger(conf)
}

// isOff - is log off
func (kl *KlynLog) isOff() bool {
	return kl.config.isOff
}

// set log off
func (kl *KlynLog) setLogOff() {
	kl.config.isOff = true
}

// writeCache - write log to cache at first
func (kl *KlynLog) writeCache(b []byte) {
	kl.cache.cacheLock.Lock()
	defer kl.cache.cacheLock.Unlock()

	kl.cache.buf.Write(b)
	return
}

// syncAndFlushCache - sync log from cache to io.writer and flush cache
func (kl *KlynLog) syncAndFlushCache() error {
	kl.cache.cacheLock.Lock()
	defer kl.cache.cacheLock.Unlock()

	if kl.cache.buf.Len() == 0 {
		return nil
	}

	cache := make([]byte, kl.cache.buf.Len())
	_, err := kl.cache.buf.Read(cache)
	if err != nil {
		return err
	}

	err = kl.getWriteAndWrite(cache)
	if err != nil {
		return err
	}

	kl.cache.buf.Reset()
	return nil
}

// getWriteAndWrite get writer and write b into
// lock when write and close writer after write
func (kl *KlynLog) getWriteAndWrite(b []byte) (err error) {
	if err = kl.getIOWriter(); err != nil {
		return
	}

	if err = kl.writeAndCloseWithLock(b); err != nil {
		fmt.Println(err)
	}

	return nil
}

// getIOWriter get log final writer and set for kl.logWriter
func (kl *KlynLog) getIOWriter() (err error) {
	day := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("%s/%s-%s.log", consts.DefaultLogDir, kl.config.Prefix, day)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}

	lw := &logWriter{
		writerLock: new(sync.RWMutex),
	}

	if kl.logWriter == nil {
		kl.logWriter = lw
	}

	kl.logWriter.writer = file

	return nil
}

// writeAndCloseWithLock - write b into kl.loggerWrier with mutex
func (kl *KlynLog) writeAndCloseWithLock(b []byte) (err error) {
	kl.logWriter.writerLock.Lock()
	defer func() {
		kl.logWriter.writer.Close()
		kl.logWriter.writerLock.Unlock()
	}()

	// append end of file
	_, err = kl.logWriter.writer.Write(b)
	return
}

// cacheLen - get cache current length of used
func (kl *KlynLog) cacheLen() int {
	kl.cache.cacheLock.RLock()
	defer kl.cache.cacheLock.RUnlock()

	return kl.cache.buf.Len()
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
	if kl.config.IsDebug {
		log.Printf(line)
	}

	kl.writeCache([]byte(line))

	return
}

// monitor
func (kl *KlynLog) monitor() {
	for {
		if kl.cacheLen() >= consts.MaxSizeOfCache {
			if err := kl.syncAndFlushCache(); err != nil {
				panic(err)
			}
		}

		select {
		case <-kl.cache.ticker.C:
			if err := kl.syncAndFlushCache(); err != nil {
				panic(err)
			}
		case e := <-kl.cache.errChan:
			fmt.Println("err:", e)
			return
		}
	}
}
