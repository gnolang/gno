package main

import (
	"fmt"

	"github.com/gnolang/gno/_test/net/http"
)

type Fromage struct {
	http.ResponseWriter
}

func main() {
	a := Fromage{}
	if a.ResponseWriter == nil {
		fmt.Println("nil")
	} else {
		fmt.Println("not nil")
	}
}

// Output:
// nil
