package main

import (
	"filetests/extern/ct1"
)

type T struct {
	m uint16
}

var t = T{1 << ct1.R}

func main() {
	println(t)
}

// Output:
// (struct{(2 uint16)} main.T)
