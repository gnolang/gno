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
// main/files/const31.gno:10:8: (const (v func()( string, string)))() (variable of type (string,string)) is not constant
