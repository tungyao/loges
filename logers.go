package loges

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
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
	isEs     bool
	file     *os.File
	fileName string
	size     int64
	writers  []io.Writer
	send     chan []byte
}

func (l *loges) trace(v ...interface{}) {

}

func (l *loges) warn(v ...interface{}) {

}

func (l *loges) error(v ...interface{}) {

}

func (l *loges) fatal(v ...interface{}) {

}
func (l *loges) request(byt []byte) {
	_, err := http.Post("", "application/json", bytes.NewReader(byt))
	if err != nil {
		l.error(err)
		return
	}
}
func (l *loges) hub(filePath string) {
	// 建立缓冲通道
	l.send = make(chan []byte, 2048)
	fs, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 766)
	if err != nil {
		log.Fatalln(err)
	}
	l.writers[0] = fs
	go func() {
		for {
			byt := <-l.send
			for _, v := range l.writers {
				v.Write(byt)
			}
		}
	}()
}

var defaultLoges *loges

func init() {
	defaultLoges = &loges{}
	defaultLoges.hub("./info.log")
}

func Println(v ...interface{}) {
	defaultLoges.trace(v)
}

func Panic(v ...interface{}) {
	defaultLoges.error(v)
}
func Warn(v ...interface{}) {
	defaultLoges.warn(v)
}
func Fatal(v ...interface{}) {
	defaultLoges.fatal(v)
}
