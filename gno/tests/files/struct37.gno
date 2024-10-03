package main

import (
	"strings"

	"github.com/gnolang/gno/_test/net/http"
)

type MyHttpClient struct {
	*http.Client
}

func main() {
	c := new(MyHttpClient)
	c.Client = new(http.Client)
	_, err := c.Get("url")
	println(strings.Contains(err.Error(), "unsupported protocol scheme"))
}

// Output:
// true
