package main

import "fmt"

func v() (string, string) {
	return "", ""
}

func main() {
	const t, v = v()
	fmt.Println(t)
}

// Error:
// main/files/const31.gno:10:8: multiple-value (const (v func()( string, string)))() (value of type [string string]) in single-value context
