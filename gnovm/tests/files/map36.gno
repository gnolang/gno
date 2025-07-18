package main

import "fmt"

var accounts = map[any]int{}

func Register() {
	var a *int
	var b *int64
	accounts[nil] = 1
	accounts[a] = 1
	accounts[b] = 1
}

func main() {
	Register()
	for k, v := range accounts {
		fmt.Printf("%v: %d\n", k, v)
	}
}

// Output:
// <nil>: 1
// <nil>: 1
// <nil>: 1
