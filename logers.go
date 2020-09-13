package loges

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
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

var (
	BasicAuth = ""
	EsUrl     = ""
)

func convert(v []interface{}) []byte {
	str := fmt.Sprintf(`{"status":"%s","datetime":"%s","pc":"%d","file":"%s","line":"%d","func":"%s","msg":"`, v[:6]...)
	for _, value := range v[6].([]interface{}) {
		str += fmt.Sprint(value) + ","
	}
	str = str[:len(str)-1]
	str += `"}`
	return []byte(str)
}
func (l *loges) trace(v ...interface{}) {
	go l.request(convert(v))
	byt := []byte(fmt.Sprintln(v))
	l.send <- byt[1 : len(byt)-2]
}

func (l *loges) warn(v ...interface{}) {
	go l.request(convert(v))
	byt := []byte(fmt.Sprintln(v))
	l.send <- byt[1 : len(byt)-2]
}

func (l *loges) error(v ...interface{}) {
	go l.request(convert(v))
	byt := []byte(fmt.Sprintln(v))
	l.send <- byt[1 : len(byt)-2]
}

func (l *loges) fatal(v ...interface{}) {
	go l.request(convert(v))
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
		res, err := c.Do(req)
		if err != nil {
			l.urlErr = true
			l.urlErrTime <- 1
			return
		}
		d, _ := ioutil.ReadAll(res.Body)
		if res.StatusCode != 200 && res.StatusCode != 201 {
			l.urlErr = true
			l.urlErrTime <- 1
			Warn(string(d))
		}
	}
}
func (l *loges) hub(filePath string) {

	// 建立缓冲通道
	l.send = make(chan []byte, 2048)
	l.writers = make([]io.Writer, 0)
	l.urlErrTime = make(chan int)
	fs, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 766)
	if err != nil {
		log.Fatalln(err)
	}
	l.writers = append(l.writers, fs)
	go func() {
		for {
			byt := <-l.send
			byt = append(byt, '\n')
			for _, v := range l.writers {
				fmt.Println(string(byt))
				go v.Write(byt)
			}
		}
	}()
	go func() {
		for {
			_ = <-l.urlErrTime
			<-time.After(time.Second * 20)
			l.urlErr = false
		}
	}()
}

var defaultLoges *loges

func init() {
	byt := base64.StdEncoding.EncodeToString([]byte(BasicAuth))
	BasicAuth = "Basic " + byt

	defaultLoges = &loges{}
	defaultLoges.hub("./info.log")
}

func Println(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.trace("info", time.Now().Format("2006-01-02T15:04:05.999999999Z"), pc, file, line, f.Name(), v)
}

func Panic(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.error("error", time.Now().Format("2006-01-02T15:04:05.999999999Z"), pc, file, line, f.Name(), v)
}
func Warn(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.warn("warn", time.Now().Format("2006-01-02T15:04:05.999999999Z"), pc, file, line, f.Name(), v)
}
func Fatal(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.fatal("fatal", time.Now().Format("2006-01-02T15:04:05.999999999Z"), pc, file, line, f.Name(), v)
}
