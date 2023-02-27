package main

import (
	"fmt"

	"github.com/gnolang/gno/_test/net/http"
)

const acceptEncoding = "Accept-Encoding"

func main() {
	opts := &http.PushOptions{
		Header: http.Header{
			acceptEncoding: []string{"gzip"},
		},
	}
	fmt.Println(opts)
}

// Output:
// &{ map[Accept-Encoding:[gzip]]}
