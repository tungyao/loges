package test

import (
	"net/http"
	"testing"

	"github.com/tungyao/loges"
)

func init() {

}
func TestLoges(t *testing.T) {
	// log.Error()
	loges.Init("./loges.log", &loges.Config{
		DevMode:  false,
		File:     false,
		EsConfig: nil,
		RabbitMq: &loges.Rabbit{
			Host:  "",
			Queue: "",
		},
	})
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		loges.Println(request.RemoteAddr, request.Method, request.URL.Path)
	})
	http.ListenAndServe(":8080", nil)
}
