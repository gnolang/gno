package main

import (
	"github.com/gnolang/gno/_test/net/http"
	"strings"
)

type S struct {
	http.Client
}

func main() {
	var s S
	if _, err := s.Get("url"); err != nil {
		println(strings.Contains(err.Error(), "unsupported protocol scheme"))
	}
	return
}

// Output:
// true
