package main

import (
	"fmt"
)

type fm map[string]interface{}

type foo struct{}

func main() {
	a := make(fm)
	a["foo"] = &foo{}
	fmt.Println(a["foo"])

	b := make(map[string]interface{})
	b["foo"] = &foo{}
	fmt.Println(b["foo"])
}

// Output:
// &{}
// &{}
