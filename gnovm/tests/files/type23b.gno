package main

import (
	"github.com/gnolang/gno/_test/net/http"
	"github.com/gnolang/gno/_test/net/http/httptest"
)

func main() {
	var v1 interface{} = 1
	var v2 interface{}
	var v3 http.ResponseWriter = httptest.NewRecorder()

	if r1, ok := v1.(string); ok {
		_ = r1
		println("unexpected")
	}
	if _, ok := v1.(string); ok {
		println("unexpected")
	}
	if r2, ok := v2.(string); ok {
		_ = r2
		println("unexpected")
	}
	if _, ok := v2.(string); ok {
		println("unexpected")
	}
	if r3, ok := v3.(http.Pusher); ok {
		_ = r3
		println("unexpected")
	}
	if _, ok := v3.(http.Pusher); ok {
		println("unexpected")
	}
	println("bye")
}

// Output:
// bye
