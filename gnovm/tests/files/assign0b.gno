package main

import (
	"fmt"
	"time"

	"filetests/extern/net/http"
)

func main() {
	http.DefaultClient.Timeout = time.Second * 10
	fmt.Println(http.DefaultClient)
	http.DefaultClient = &http.Client{}
	fmt.Println(http.DefaultClient)
}

// Error:
// cannot directly modify readonly tainted object (w/o method): (const (ref(filetests/extern/net/http) package{})).DefaultClient.Timeout
