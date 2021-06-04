package main

import (
	"github.com/gnolang/gno/_test/net"
	"strings"
)

type ipNetValue net.IPNet

func (ipnet *ipNetValue) Set(value string) error {
	_, n, err := net.ParseCIDR(strings.TrimSpace(value))
	if err != nil {
		return err
	}
	*ipnet = ipNetValue(*n)
	return nil
}

func main() {
	v := ipNetValue{}
	println(v)
}

// Output:
// struct{(undefined),(undefined)}
