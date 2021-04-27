package main

import (
	"github.com/gnolang/gno/_test/context"
	"github.com/gnolang/gno/_test/net"
)

var lookupHost = net.DefaultResolver.LookupHost

func main() {
	res, err := lookupHost(context.Background(), "localhost")
	println(len(res) > 0, err == nil)
}

// Output:
// true true
