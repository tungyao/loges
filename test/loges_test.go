package test

import (
	"../../loges2"
	"net/http"
	"testing"
)

func init() {
	loges.Init("", "", "")
}
func TestLoges(t *testing.T) {
	// log.Error()
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		loges.Println(request.RemoteAddr, request.Method, request.URL.Path)
	})
	http.ListenAndServe(":8000", nil)
}
