package main

import (
	"strings"

	"filetests/extern/net"
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
// (struct{(nil filetests/extern/net.IP),(nil filetests/extern/net.IPMask)} main.ipNetValue)
