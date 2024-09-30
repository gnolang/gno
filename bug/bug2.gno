package main

import (
	"fmt"
	"errors"
	"io"
	"time"
)

type Header map[string][]string

type Values map[string][]string

type Request struct {
	Method           string
	Proto            string // "HTTP/1.0"
	ProtoMajor       int    // 1
	ProtoMinor       int    // 0
	Header           Header
	ContentLength    int64
	TransferEncoding []string
	Close            bool
	Host             string
}

type Response struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.0"
	ProtoMajor int    // e.g. 1
	ProtoMinor int    // e.g. 0
}

type extendedRequest struct {
	Request2 Request

	Data string
}

func main() {
	r := extendedRequest{}
	req := &r.Request2

	fmt.Println(r)
	fmt.Println(req)
}

// Output:
// {{  0 0 map[] <nil> 0 [] false  map[] map[] map[]   <nil>} }
// &{  0 0 map[] <nil> 0 [] false  map[] map[] map[]   <nil>}
