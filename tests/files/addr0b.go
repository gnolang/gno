package main

import (
	"fmt"

	"github.com/gnolang/gno/_test/net/http"
)

type extendedRequest struct {
	Request http.Request

	Data string
}

func main() {
	r := extendedRequest{}
	req := &r.Request

	fmt.Println(r)
	fmt.Println(req)
}

// Output:
// {{  0 0 map[] <nil> 0 [] false  map[] map[] map[]   <nil>} }
// &{  0 0 map[] <nil> 0 [] false  map[] map[] map[]   <nil>}
