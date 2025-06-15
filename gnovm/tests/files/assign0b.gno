package main

import (
	"fmt"
	"time"

	"github.com/gnolang/gno/_test/net/http"
)

func main() {
	http.DefaultClient.Timeout = time.Second * 10
	fmt.Println(http.DefaultClient)
	http.DefaultClient = &http.Client{}
	fmt.Println(http.DefaultClient)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): (const (ref(github.com/gnolang/gno/_test/net/http) package{})).DefaultClient.Timeout
