package main

import (
	"fmt"
)

type NRGBA64 struct {
	R, G, B, A uint16
}

func main() {
	c := NRGBA64{1, 1, 1, 1}

	fmt.Println(c)
}

// Output:
// {1 1 1 1}
