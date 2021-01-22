package test

import (
	"../../loges2"
	"net/http"
	"testing"
)

func init() {
	loges.Init("", "", "./loges.log", false, &loges.Config{
		RabbitMq: loges.Rabbit{
			Host:  "amqp://tungyao:JLASlj12jlias@106.52.170.25:5672/admin",
			Queue: "loges_test",
		},
	})
}
func TestLoges(t *testing.T) {
	// log.Error()
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		loges.Panic(request.RemoteAddr, request.Method, request.URL.Path)
	})
	http.ListenAndServe(":8000", nil)
}
