// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

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

const (
	// FlushModeEveryLog -  flush cache to disk each log
	FlushModeEveryLog = iota
	// FlushModeByDuration - flush cache to disk with every duration
	FlushModeByDuration
	// FlushModeBySize - flush cache to disk only when cache larger then size setted
	FlushModeBySize
)

// KlynLog - implement Logger and provide cache
type KlynLog struct {
	config    *LoggerConfig
	logWriter *logWriter // log final destination
	cache     *logCache  // log temp cache
}

// LoggerConfig - logger config
type LoggerConfig struct {
	isOff bool

	FlushMode int // flush dick mode
	IsDebug   bool
	Prefix    string
}

type logCache struct {
	buf       *bytes.Buffer
	ticker    *time.Ticker
	forceSync chan bool
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
		forceSync: make(chan bool, 1),
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
		Prefix:    "KLYN",
		FlushMode: FlushModeByDuration,
	}

	return NewLogger(conf)
}

// isOff - is log off
func (kl *KlynLog) isOff() bool {
	return kl.config.isOff
}

// isFlushEveryLog -  is flush mode is FlushModeEveryLog
func (kl *KlynLog) isFlushEveryLog() bool {
	return kl.config.FlushMode == FlushModeEveryLog
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

	// already locked so no need to call `cacheLen()`
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

	if kl.logWriter == nil {
		kl.logWriter = &logWriter{
			writerLock: new(sync.RWMutex),
		}
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
func (kl *KlynLog) cacheLen() (n int) {
	kl.cache.cacheLock.RLock()
	defer kl.cache.cacheLock.RUnlock()

	n = kl.cache.buf.Len()
	return
}

// Trace - trace level log
func (kl *KlynLog) Trace(j interface{}) {
	kl.log(LoggerLevelTrace, j)
}

// Debug - debug level log
func (kl *KlynLog) Debug(j interface{}) {
	kl.log(LoggerLevelDebug, j)
}

// Info - info level log
func (kl *KlynLog) Info(j interface{}) {
	kl.log(LoggerLevelInfo, j)
}

// Warn - warn level info
func (kl *KlynLog) Warn(j interface{}) {
	kl.log(LoggerLevelWarn, j)
}

// Error - error level info
func (kl *KlynLog) Error(j interface{}) {
	kl.log(LoggerLevelError, j)
}

// Fatal - fatal level info
func (kl *KlynLog) Fatal(j interface{}) {
	kl.log(LoggerLevelFatal, j)
}

// Any - custom level log
// level should be valid
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

	if kl.isFlushEveryLog() {
		kl.syncAndFlushCache()
	}

	return
}

// monitor - monitoring forceSync chan, flush cache once channel receive value
func (kl *KlynLog) monitor() {
	// only monitor forceSync chan when need monitoring
	switch kl.config.FlushMode {
	case FlushModeBySize:
		go kl.sizeMonitor()
	case FlushModeByDuration:
		go kl.durationMonitor()
	// other mode no need to monitoring forceSync chan
	default:
		return
	}

	for {
		select {
		case <-kl.cache.forceSync:
			if err := kl.syncAndFlushCache(); err != nil {
				panic(err)
			}
		case e := <-kl.cache.errChan:
			fmt.Println("err:", e)
			return
		}
	}
}

// sizeMonitor - check cache size every 10 millisecond.
// Send a value to forceSync channel when cache size large then MaxSizeOfCache
func (kl *KlynLog) sizeMonitor() {
	for {
		if kl.cacheLen() >= consts.MaxSizeOfCache {
			kl.cache.forceSync <- true
		} else {
			time.Sleep(time.Millisecond * 10)
		}
	}
}

// durationMonitor - send value to forceSync when every tick
func (kl *KlynLog) durationMonitor() {
	for {
		select {
		case <-kl.cache.ticker.C:
			kl.cache.forceSync <- true
		}
	}
}
