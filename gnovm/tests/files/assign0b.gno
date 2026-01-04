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
// main/assign0b.gno:13:2-20: cannot directly mutate filetests/extern/net/http.DefaultClient from main
