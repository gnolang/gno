package main

import (
	"strings"

	"github.com/gnolang/gno/_test/net"
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
// struct{(nil github.com/gnolang/gno/_test/net.IP),(nil github.com/gnolang/gno/_test/net.IPMask)}
