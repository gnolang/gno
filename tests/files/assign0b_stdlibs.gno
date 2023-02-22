package main

import (
	"fmt"
	"github.com/gnolang/gno/_test/net/http"
	"time"
)

func main() {
	http.DefaultClient.Timeout = time.Second * 10
	fmt.Println(http.DefaultClient)
	http.DefaultClient = &http.Client{}
	fmt.Println(http.DefaultClient)
}

// Output:
// &{<nil> <nil> 10000000000}
// &{<nil> <nil> 0}
