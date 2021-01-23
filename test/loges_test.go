package test

import (
	"net/http"
	"testing"

	"../../loges"
)

func init() {
	loges.Init("", "", "./loges.log", false, &loges.Config{
		DevMode:false,
		RabbitMq: loges.Rabbit{
			Host:  "",
			Queue: "",
		},
	})
}
func TestLoges(t *testing.T) {
	// log.Error()
	loges.Println("hello world")
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		loges.Panic(request.RemoteAddr, request.Method, request.URL.Path)
	})
	http.ListenAndServe(":8000", nil)
}
