package loges

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

// 日志 输出至 ES , FILE

// 日志调用API
type logger interface {
	trace(v ...interface{})
	warn(v ...interface{})
	error(v ...interface{})
	fatal(v ...interface{})
}

// 日志记录器
// isEs 是否开启es
// file 文件句柄
// fileName 日志文件位置
// size 日志文件大小
// writers 向外写出器
type loges struct {
	logger
	sync.Mutex
	isEs       bool
	file       *os.File
	fileName   string
	size       int64
	writers    []io.Writer
	send       chan []byte
	urlErr     bool
	urlErrTime chan int
}

const (
	BasicAuth = "Basic 123123"
	EsUrl     = "http://123123"
)

func (l *loges) trace(v ...interface{}) {
	byt := []byte(fmt.Sprintln(v))
	l.send <- byt[1 : len(byt)-2]
}

func (l *loges) warn(v ...interface{}) {
	byt := []byte(fmt.Sprintln(v))
	l.send <- byt[1 : len(byt)-2]
}

func (l *loges) error(v ...interface{}) {
	byt := []byte(fmt.Sprintln(v))
	l.send <- byt[1 : len(byt)-2]
}

func (l *loges) fatal(v ...interface{}) {
	byt := []byte(fmt.Sprintln(v))
	l.send <- byt[1 : len(byt)-2]
}
func (l *loges) request(byt []byte) {
	if !l.urlErr {
		c := http.Client{}
		req, err := http.NewRequest("POST", EsUrl, bytes.NewReader(byt))
		if err != nil {
			l.urlErr = true
			l.urlErrTime <- 1
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", BasicAuth)
		c.Do(req)
	}
}
func (l *loges) hub(filePath string) {
	// 建立缓冲通道
	l.send = make(chan []byte, 2048)
	l.writers = make([]io.Writer, 0)
	fs, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 766)
	if err != nil {
		log.Fatalln(err)
	}
	l.writers = append(l.writers, fs)
	go func() {
		for {
			byt := <-l.send
			for _, v := range l.writers {
				v.Write(byt)
			}
		}
	}()
	go func() {
		for {
			_ = <-l.urlErrTime
			<-time.After(time.Second * 30)
			l.urlErr = false
		}
	}()
}

var defaultLoges *loges

func init() {
	defaultLoges = &loges{}
	defaultLoges.hub("./info.log")
}

func Println(v ...interface{}) {
	pc, file, line, ok := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.trace("info", time.Now().Format("2006-01-02 15:04:05"), pc, file, line, ok, f.Name(), v)
}

func Panic(v ...interface{}) {
	pc, file, line, ok := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.error("error", time.Now().Format("2006-01-02 15:04:05"), pc, file, line, ok, f.Name(), v)
}
func Warn(v ...interface{}) {
	pc, file, line, ok := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.warn("warn", time.Now().Format("2006-01-02 15:04:05"), pc, file, line, ok, f.Name(), v)
}
func Fatal(v ...interface{}) {
	pc, file, line, ok := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.fatal("fatal", time.Now().Format("2006-01-02 15:04:05"), pc, file, line, ok, f.Name(), v)
}
