package loges

import (
	"encoding/base64"
	"fmt"
	"io"
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
	isEs       bool
	file       *os.File
	fileName   string
	size       int64
	writers    []io.Writer
	send       chan []byte
	urlErr     bool
	urlErrTime chan int
	config     *Config
}

type Config struct {
	RabbitMq Rabbit
	DevMode  bool // is dev mode,not print log
}
type Rabbit struct {
	Host  string
	Queue string
}

// 在这里配置 密码 和 url
var (
	BasicAuth = ""
	EsUrl     = ""
)

// 增加初始化方法
func Init(esUrl, basicAuth, logPath string, isEs bool, config *Config) *loges {
	BasicAuth = basicAuth
	EsUrl = esUrl
	byt := base64.StdEncoding.EncodeToString([]byte(BasicAuth))
	BasicAuth = "Basic " + byt
	defaultLoges = &loges{
		isEs:   isEs,
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
func (l *loges) trace(v ...interface{}) {
	if l.isEs {
		go l.request(convert(v))
	}
	byt := []byte(fmt.Sprintln(v))
	l.send <- byt[1 : len(byt)-2]
}

func (l *loges) warn(v ...interface{}) {
	if l.isEs {
		go l.request(convert(v))
	}
	byt := []byte(fmt.Sprintln(v))
	l.send <- byt[1 : len(byt)-2]
}

func (l *loges) error(v ...interface{}) {
	if l.isEs {
		go l.request(convert(v))
	}
	byt := []byte(fmt.Sprintln(v))
	l.send <- byt[1 : len(byt)-2]
	panic(v)
}

func (l *loges) fatal(v ...interface{}) {
	if l.isEs {
		go l.request(convert(v))
	}
	byt := []byte(fmt.Sprintln(v))
	l.send <- byt[1 : len(byt)-2]
	panic(v)
}
func (l *loges) request(byt string) {
	if !l.urlErr {
		c := http.Client{}
		req, err := http.NewRequest("POST", EsUrl, strings.NewReader(byt))
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

	// build channel
	l.send = make(chan []byte, 2048)
	l.writers = make([]io.Writer, 0)
	l.urlErrTime = make(chan int)
	fs, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 766)
	if err != nil {
		log.Fatalln(err)
	}
	l.writers = append(l.writers, fs)
	if l.config != nil {
		// rabbitmq connect
		conn, err := amqp.Dial(l.config.RabbitMq.Host)
		if err != nil {
			log.Fatal("connect amqp failed")
		}
		ch, err := conn.Channel()
		if err != nil {
			log.Fatal("connect channel failed")
		}
		q, err := ch.QueueDeclare(
			l.config.RabbitMq.Queue, // name
			false,                   // durable
			false,                   // delete when unused
			false,                   // exclusive
			false,                   // no-wait
			nil,                     // arguments
		)
		outer := &MqOuter{
			Queue: q,
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

var defaultLoges *loges

func Println(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.trace("info", time.Now().Format(time.RFC3339), pc, file, line, f.Name(), v)
}

func Panic(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.error("error", time.Now().Format(time.RFC3339), pc, file, line, f.Name(), v)
}
func Warn(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.warn("warn", time.Now().Format(time.RFC3339), pc, file, line, f.Name(), v)
}
func Fatal(v ...interface{}) {
	pc, file, line, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	defaultLoges.fatal("fatal", time.Now().Format(time.RFC3339), pc, file, line, f.Name(), v)
}

// mqOuter
type MqOuter struct {
	io.ByteWriter
	Amqp  *amqp.Channel
	Queue amqp.Queue
}

func (m *MqOuter) Write(p []byte) (n int, err error) {
	err = m.Amqp.Publish(
		"",
		m.Queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        p,
		},
	)
	return len(p), err
}
