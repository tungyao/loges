package test

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	log "../../loges"
)

func TestLoges(t *testing.T) {
	// log.Error()
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Println(request.RemoteAddr, request.Method, request.URL.Path)
	})
	http.ListenAndServe(":8000", nil)
}
func TestRef(t *testing.T) {
	v := []byte("123")
	// v:="123"
	types := reflect.TypeOf(v)
	fmt.Println(types.String())
	switch types.Name() {
	case "string":
		fmt.Println(123)
	case "[]uint8":

	}
}
func TestName(t *testing.T) {
	t.Log(time.Now().Format("2006-01-02T15:04:05.999999999Z"))
}
