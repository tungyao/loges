package loges

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// 日志调用API
type logger interface {
	trace(v ...interface{})
	warn(v ...interface{})
	error(v ...interface{})
	fatal(v ...interface{})
}

// isEs enable es
// file file
// fileName log file storage path
// size file size
// writers io.writer
type loges struct {
	logger
	sync.Mutex
	file       *os.File
	fileName   string
	size       int64
	writers    []LogesWriter
	send       chan *DataStruct
	urlErr     bool
	urlErrTime chan int
	config     *Config
}

type Config struct {
	EsConfig *Es
	RabbitMq *Rabbit
	File     bool
	DevMode  bool // is dev mode,not print log
}
type Rabbit struct {
	Host  string
	Queue string
}
type Es struct {
	Host      string
	BasicAuth string
}

// 增加初始化方法
func Init(logPath string, config *Config) *loges {
	defaultLoges = &loges{
		config: config,
	}
	defaultLoges.hub(logPath)
	return defaultLoges
}

func convert(v []interface{}) string {
	str := fmt.Sprintf(`{"status":"%s","datetime":"%s","pc":"%d","file":"%s","line":"%d","func":"%s","msg":"`, v[:6]...)
	// s := ""
	for _, value := range v[6].([]interface{}) {
		str += fmt.Sprint(value) + ","
	}
	str = str[:len(str)-1]
	str = strings.Replace(str, string(uint8(9)), "", -1)
	return str + `"}`
}
func (l *loges) trace(dataStruct *DataStruct) {
	l.send <- dataStruct
}

func (l *loges) warn(dataStruct *DataStruct) {
	l.send <- dataStruct
}

func (l *loges) error(dataStruct *DataStruct) {
	l.send <- dataStruct
	panic(dataStruct.Msg)
}

func (l *loges) fatal(dataStruct *DataStruct) {
	l.send <- dataStruct
	panic(dataStruct.Msg)
}
func (es *EsOuter) request(byt []byte) {
	if !es.urlErr {
		c := http.Client{}
		req, err := http.NewRequest("POST", es.Host, bytes.NewReader(byt))
		if err != nil {
			es.urlErr = true
			es.urlErrTime <- 1
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", es.BasicAuth)
		res, err := c.Do(req)
		if err != nil {
			es.urlErr = true
			es.urlErrTime <- 1
			return
		}
		d, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(d))
		if res.StatusCode != 200 && res.StatusCode != 201 {
			es.urlErr = true
			es.urlErrTime <- 1
			Warn(string(d))
		}
	}
}

var defaultLoges *loges

func (l *loges) hub(filePath string) {

	// build channel
	l.send = make(chan *DataStruct, 2048)
	l.writers = make([]LogesWriter, 0)
	l.urlErrTime = make(chan int)

	// append local file
	if l.config.File {
		fs, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 766)
		if err != nil {
			log.Fatalln(err)
		}
		// append file
		l.writers = append(l.writers, &FileOuter{
			fs: fs,
		})
	}

	// append es
	if l.config.EsConfig != nil {
		l.config.EsConfig.BasicAuth = "Basic " + base64.StdEncoding.EncodeToString([]byte(l.config.EsConfig.BasicAuth))
		esOuters := &EsOuter{
			Host:       l.config.EsConfig.Host,
			BasicAuth:  l.config.EsConfig.BasicAuth,
			urlErr:     l.urlErr,
			urlErrTime: l.urlErrTime,
		}
		l.writers = append(l.writers, esOuters)
	}
	// append rabbit
	if l.config.RabbitMq != nil {
		// rabbitmq connect
		conn, err := amqp.Dial(l.config.RabbitMq.Host)
		if err != nil {
			fmt.Println("connect amqp failed")
		}
		ch, err := conn.Channel()
		if err != nil {
			fmt.Println("connect channel failed")
		}

		outer := &MqOuter{
			Queue: l.config.RabbitMq.Queue,
			Amqp:  ch,
		}
		l.writers = append(l.writers, outer)
	}

	go func() {
		for {
			byt := <-l.send
			// start dev mode
			for i, v := range l.writers {
				if l.config.DevMode && i != 0 {
					continue
				}
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

func Println(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.trace(&DataStruct{
		Status:   "info",
		DateTime: time.Now().Format(time.RFC3339),
		Pc:       int64(pc),
		File:     file,
		Line:     line,
		Func:     f.Name(),
		Msg:      fmt.Sprint(v),
	})
}

func Panic(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.error(&DataStruct{
		Status:   "error",
		DateTime: time.Now().Format(time.RFC3339),
		Pc:       int64(pc),
		File:     file,
		Line:     line,
		Func:     f.Name(),
		Msg:      fmt.Sprint(v),
	})
}
func Warn(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.warn(&DataStruct{
		Status:   "warn",
		DateTime: time.Now().Format(time.RFC3339),
		Pc:       int64(pc),
		File:     file,
		Line:     line,
		Func:     f.Name(),
		Msg:      fmt.Sprint(v),
	})
}
func Fatal(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.fatal(&DataStruct{
		Status:   "fatal",
		DateTime: time.Now().Format(time.RFC3339),
		Pc:       int64(pc),
		File:     file,
		Line:     line,
		Func:     f.Name(),
		Msg:      fmt.Sprint(v),
	})
}
