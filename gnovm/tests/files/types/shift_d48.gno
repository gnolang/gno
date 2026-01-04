package main

import (
	"filetests/extern/net/http"
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
// (struct{(struct{( string),( string),(0 int),(0 int),(nil filetests/extern/net/http.Header),(undefined),(0 int64),(nil []string),(false bool),( string),(nil filetests/extern/net/http.Values),(nil filetests/extern/net/http.Values),(nil filetests/extern/net/http.Header),( string),( string),(nil *filetests/extern/net/http.Response)} filetests/extern/net/http.Request),( string)} main.extendedRequest)
