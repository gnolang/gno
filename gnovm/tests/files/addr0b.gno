package main

import (
	"github.com/gnolang/gno/_test/net/http"
)

type extendedRequest struct {
	Request http.Request

	Data string
}

func main() {
	r := extendedRequest{}
	// req := &r.Request

	println(r)
	// XXX removed temporarily until recursion detection implemented for sprintString().
	// println(req)
}

// Output:
// (struct{(struct{( string),( string),(0 int),(0 int),(nil github.com/gnolang/gno/_test/net/http.Header),(undefined),(0 int64),(nil []string),(false bool),( string),(nil github.com/gnolang/gno/_test/net/http.Values),(nil github.com/gnolang/gno/_test/net/http.Values),(nil github.com/gnolang/gno/_test/net/http.Header),( string),( string),(nil *github.com/gnolang/gno/_test/net/http.Response)} github.com/gnolang/gno/_test/net/http.Request),( string)} main.extendedRequest)
