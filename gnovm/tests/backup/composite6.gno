package main

import (
	"fmt"

	"github.com/gnolang/gno/_test/ct1"
)

type T struct {
	m uint16
}

var t = T{1 << ct1.R}

func main() {
	fmt.Println(t)
}

// Output:
// {2}
