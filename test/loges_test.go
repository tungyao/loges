package test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	log "../../loges2"
)

func TestLoges(t *testing.T) {
	// log.Error()
	log.Println(errors.New("123"))
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
