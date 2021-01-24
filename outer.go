package loges

import (
	"bytes"
	"fmt"
	"os"

	"github.com/streadway/amqp"
)

// fileOuter
type FileOuter struct {
	LogesWriter
	fs *os.File
}

func (m *FileOuter) Write(p []interface{}) (n int, err error) {
	b := fmt.Sprint(p)
	m.fs.Write(bytes.NewBufferString(b[1:len(b)-1] + "\n").Bytes())
	m.fs.Sync()
	return len(p), err
}

// mqOuter
type MqOuter struct {
	LogesWriter
	Amqp  *amqp.Channel
	Queue amqp.Queue
}

func (m *MqOuter) Write(p []interface{}) (n int, err error) {
	err = m.Amqp.Publish(
		"",
		m.Queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(fmt.Sprintln(p)),
		},
	)
	return len(p), err
}

// es outer
type EsOuter struct {
	LogesWriter
	urlErr     bool
	urlErrTime chan int
	Host       string
	BasicAuth  string
}

func (es *EsOuter) Write(p []interface{}) (n int, err error) {
	es.request([]byte(convert(p)))
	return len(p), err
}
