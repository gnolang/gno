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
// main/files/const31.gno:10:8: (const (v func() (string, string)))() (variable of type (string,string)) is not constant

// TypeCheckError:
// main/files/const31.gno:10:11: missing init expr for v; main/files/const31.gno:10:15: multiple-value v() (value of type (string, string)) in single-value context
