package test

import (
	"fmt"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	fmt.Println(strings.Index("'trade_state':'SUCCESS',", "trade_state"))
	ss := "'trade_state':'SUCCESS'"
	fmt.Println(ss[1+14:])
}
