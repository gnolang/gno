package main

import "fmt"

type s struct {
	a *int
}

var mp = map[s]int{}

func main() {
	mp[s{}] = 1
	for k, v := range mp {
		fmt.Printf("%v: %d\n", k, v)
	}
}

// Output:
// {<nil>}: 1
