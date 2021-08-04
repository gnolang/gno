package main

import (
	"fmt"
	"github.com/gnolang/gno/_test/net/http"
	"github.com/gnolang/gno/_test/net/http/httptest"
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

	var v interface{} = 1
	println(v.(string))
}

func assertNil() {
	defer func() {
		r := recover()
		fmt.Println(r)
	}()

	var v interface{}
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
// interface{} is not of type string
// *github.com/gnolang/gno/_test/net/http/httptest.ResponseRecorder doesn't implement interface{Push#581DCF94C557261E8D3369A703097601AF0E85A7}
