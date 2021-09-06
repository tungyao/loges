package loges

import (
	"os"

	json "github.com/json-iterator/go"
	"github.com/streadway/amqp"
)

// fileOuter
type FileOuter struct {
	LogesWriter
	fs *os.File
}

func (m *FileOuter) Write(dataStruct *DataStruct) (n int, err error) {
	d, err := json.Marshal(dataStruct)
	m.fs.Write(d)
	m.fs.Write([]byte("\r\n"))
	m.fs.Sync()
	return 0, err
}

// mqOuter
type MqOuter struct {
	LogesWriter
	Amqp  *amqp.Channel
	Queue string
}

func (m *MqOuter) Write(dataStruct *DataStruct) (n int, err error) {
	d, err := json.Marshal(dataStruct)
	q, err := m.Amqp.QueueDeclare(
		m.Queue, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	err = m.Amqp.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        d,
		},
	)
	return 0, err
}

// es outer
type EsOuter struct {
	LogesWriter
	urlErr     bool
	urlErrTime chan int
	Host       string
	BasicAuth  string
}

func (es *EsOuter) Write(dataStruct *DataStruct) (n int, err error) {
	d, err := json.Marshal(dataStruct)
	es.request(d)
	return 0, err
}
