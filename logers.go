package loges

import (
	"io"
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
}

func (l *loges) trace(v ...interface{}) {

}

func (l *loges) warn(v ...interface{}) {

}

func (l *loges) error(v ...interface{}) {

}

func (l *loges) fatal(v ...interface{}) {

}

var defaultLoges *loges

func init() {
	defaultLoges = &loges{}
}

func Println(v ...interface{}) {
	defaultLoges.trace(v)
}

func Panic(v ...interface{}) {

}
func Warn(v ...interface{}) {

}
func Fatal(v ...interface{}) {

}
