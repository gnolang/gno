package main

import (
	"fmt"

	"filetests/extern/net/http"
)

type Fromage struct {
	http.Server
}

func main() {
	a := Fromage{}
	fmt.Println(a.Server.WriteTimeout)
}

// Output:
// 0s
