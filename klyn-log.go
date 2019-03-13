// Copyright 2018 Yusan Kurban. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package klynlog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"git.yusank.cn/yusank/klyn-log/consts"
	"git.yusank.cn/yusank/klyn-log/utils"

	"github.com/json-iterator/go"
)

var (
	JSON = jsoniter.Config{
		EscapeHTML:  true,
		SortMapKeys: true,
	}.Froze()
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

// NewLogger return Logger
func NewLogger(l *LoggerConfig) Logger {
	cache := &logCache{
		buf:       new(bytes.Buffer),
		cacheLock: new(sync.RWMutex),
		ticker:    time.NewTicker(consts.DefaultTickerDuration),
		syncChan:  make(chan bool, 0),
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

	c := make(chan os.Signal)
	// 监听信号量
	// SIGHUP: 终端结束进程(终端连接断开)
	// SIGTERM: 结束程序
	// SIGINT: Ctrl+ c 操作
	// SIGQUIT: Ctrl+ / 操作
	// SIGUSR1, SIGUSR2 用户保留
	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT,
		syscall.SIGUSR1, syscall.SIGUSR2)

	go func() {
		for {
			// 如捕捉到监听的信号，将内存中的日志写入文件
			s := <-c
			log.Println("catch signal:", s.String())
			_ = logger.syncAndFlushCache()
			switch s {
			// 如果为退出信号 则安全退出
			case syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
				os.Exit(0)
			// 可以通过给进程发送 syscall.SIGUSR1, syscall.SIGUSR2 信号来，强制将缓存中的日志写入文件
			default:
				log.Fatal(s.String())
			}
		}
	}()

	return logger
}

// DefaultLogger - get default logger
func DefaultLogger() Logger {
	conf := &LoggerConfig{
		Prefix:    "KLYN",
		FlushMode: consts.FlushModeByDuration,
		IsDebug:   true,
	}

	return NewLogger(conf)
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
func (kl *KlynLog) Any(level Level, j interface{}) {
	kl.log(level, j)
}

// OFF - off all level log
func (kl *KlynLog) OFF() {
	kl.setLogOff()
}

func (kl *KlynLog) log(l Level, j interface{}) {
	if kl.isOff() || j == nil {
		return
	}

	b, _ := JSON.Marshal(j)

	line := fmt.Sprintf("[%s] | LEVEL:%s | message:%s\n", kl.config.Prefix, l.String(), string(b))
	if kl.config.IsDebug {
		log.Printf(line)
	}

	if kl.isFlushEveryLog() {
		// if flush every log to io, then no need to write to cache
		_ = kl.writeToIO([]byte(line))
		return
	}

	if err := kl.cache.write([]byte(line)); err != nil {
		log.Fatal(err)
	}

	return
}

// isOff - is log off
func (kl *KlynLog) isOff() bool {
	return kl.config.isOff
}

// isFlushEveryLog -  is flush mode is FlushModeEveryLog
func (kl *KlynLog) isFlushEveryLog() bool {
	return kl.config.FlushMode == consts.FlushModeEveryLog
}

// set log off
func (kl *KlynLog) setLogOff() {
	kl.config.isOff = true
}

func (kl *KlynLog) isWriterClosed() bool {
	return kl.logWriter == nil || kl.logWriter.writer == nil
}

func (kl *KlynLog) isWriterValid() bool {
	return kl.logWriter != nil && kl.logWriter.writer != nil
}

// syncAndFlushCache - sync log from cache to io.writer and flush cache
func (kl *KlynLog) syncAndFlushCache() error {
	// already locked so no need to call `cacheLen()`
	if kl.cache.length() == 0 {
		return nil
	}

	cache, err := kl.cache.popCache()
	if err != nil {
		return err
	}

	err = kl.writeToIO(cache)
	if err != nil {
		return err
	}

	return nil
}

// writeToIO get writer and write b into
// lock when write and close writer after write
func (kl *KlynLog) writeToIO(b []byte) (err error) {
	if err = kl.getIOWriter(); err != nil {
		return
	}

	if err = kl.logWriter.write(b); err != nil {
		fmt.Println(err)
	}

	return nil
}

// getIOWriter get log final writer and set for kl.logWriter
func (kl *KlynLog) getIOWriter() (err error) {
	// io writer still valid
	if kl.isWriterValid() {
		return
	}

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

// MaintainIOWriter - maintain kl io writer, in case opened and closed too frequently.
// only run flush every log mode
func (kl *KlynLog) MaintainIOWriter() {
	var ts int64
	for {
		if kl.isWriterClosed() {
			goto sleep
		}

		kl.logWriter.writerLock.RLock()

		ts = time.Now().UnixNano()
		if kl.logWriter.lastWrite != 0 && ts-kl.logWriter.lastWrite > int64(1*time.Second) && kl.logWriter.writer != nil {
			if err := kl.logWriter.close(); err != nil {
				log.Fatal(err)
			}
		}
		kl.logWriter.writerLock.RUnlock()

	sleep:
		time.Sleep(100 * time.Millisecond)
	}
}

// cacheLen - get cache current length of used
func (kl *KlynLog) cacheLen() (n int) {
	kl.cache.cacheLock.RLock()
	defer kl.cache.cacheLock.RUnlock()

	n = kl.cache.buf.Len()
	return
}

// monitor - monitoring syncChan chan, flush cache once channel receive value
func (kl *KlynLog) monitor() {
	// only monitor syncChan chan when need monitoring
	switch kl.config.FlushMode {
	case consts.FlushModeBySize:
		go kl.sizeMonitor()
	case consts.FlushModeByDuration:
		go kl.durationMonitor()
	case consts.FlushModeEveryLog:
		go kl.MaintainIOWriter()
	// other mode no need to monitoring syncChan chan
	default:
		return
	}

	for {
		select {
		case <-kl.cache.syncChan:
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
// Send a value to syncChan channel when cache size large then MaxSizeOfCache
func (kl *KlynLog) sizeMonitor() {
	for {
		if kl.cacheLen() >= consts.MaxSizeOfCache {
			kl.cache.syncChan <- true
		} else {
			time.Sleep(time.Millisecond * 10)
		}
	}
}

// durationMonitor - send value to syncChan when every tick
func (kl *KlynLog) durationMonitor() {
	for {
		select {
		case <-kl.cache.ticker.C:
			kl.cache.syncChan <- true
		}
	}
}

type logCache struct {
	buf       *bytes.Buffer
	ticker    *time.Ticker
	syncChan  chan bool
	errChan   chan error
	cacheLock *sync.RWMutex
}

// writeCache - write log to cache at first
func (lc *logCache) write(b []byte) error {
	lc.cacheLock.Lock()
	defer lc.cacheLock.Unlock()

	_, err := lc.buf.Write(b)
	return err
}

func (lc *logCache) length() int {
	lc.cacheLock.RLock()
	defer lc.cacheLock.RUnlock()

	l := lc.buf.Len()
	return l
}

// read and reset cache
func (lc *logCache) popCache() (p []byte, err error) {
	lc.cacheLock.Lock()
	defer lc.cacheLock.Unlock()

	p = make([]byte, lc.buf.Len())
	_, err = lc.buf.Read(p)
	if err != nil {
		return
	}

	lc.buf.Reset()
	return
}

type logWriter struct {
	lastWrite  int64 // last write time stamp
	writer     io.WriteCloser
	writerLock *sync.RWMutex
}

func (lw *logWriter) write(b []byte) error {
	lw.writerLock.Lock()
	defer lw.writerLock.Unlock()

	_, err := lw.writer.Write(b)
	lw.lastWrite = time.Now().UnixNano()
	return err
}

func (lw *logWriter) close() error {
	lw.writerLock.Lock()
	defer lw.writerLock.Unlock()

	err := lw.writer.Close()
	return err
}
