package main

import (
	"fmt"
)

func main() {
	a := []func(string){bar}
	b := a[0]
	b("bar")
}

func bar(a string) {
	fmt.Println(a)
}

// Output:
// bar
