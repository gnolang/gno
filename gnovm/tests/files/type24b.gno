package main

import (
	"fmt"

	"filetests/extern/net/http"
	"filetests/extern/net/http/httptest"
)

func main() {
	assertInt()
	assertNil()
	assertValue()
}

func assertInt() {
	defer func() {
		r := recover()
		fmt.Println(r)
	}()

	var v any = 1
	println(v.(string))
}

func assertNil() {
	defer func() {
		r := recover()
		fmt.Println(r)
	}()

	var v any
	println(v.(string))
}

func assertValue() {
	defer func() {
		r := recover()
		fmt.Println(r)
	}()

	var v http.ResponseWriter = httptest.NewRecorder()
	println(v.(http.Pusher))
}

// Output:
// int is not of type string
// nil is not of type string
// *filetests/extern/net/http/httptest.ResponseRecorder doesn't implement interface {Push func(string, *filetests/extern/net/http.PushOptions) .uverse.error} (missing method Push)
