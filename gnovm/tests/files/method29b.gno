package main

import (
	"filetests/extern/context"
	"filetests/extern/net"
)

var lookupHost = net.DefaultResolver.LookupHost

func main() {
	res, err := lookupHost(context.Background(), "localhost")
	println(len(res) > 0, err == nil)
}

// Output:
// true true
