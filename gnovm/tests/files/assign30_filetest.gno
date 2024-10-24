package main

import "fmt"

func t() (int, string) { return 1, "str" }

func main() {
	var v interface{} = 1

	i, x := v.(int), 2
}
