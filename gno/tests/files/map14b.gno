package main

import "fmt"

var m = map[string]int{"foo": 1}

func f(s string) bool { return m[s] > 0 }

func main() {
	fmt.Println(f("foo"), f("bar"))
}

// Output:
// true false
