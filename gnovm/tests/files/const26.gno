package main

import "fmt"

func v() string {
	return ""
}

func main() {
	const t = v()
	fmt.Println(t)
}

// Error:
// main/files/const26.gno:10:8: v<VPBlock(3,0)>() (value of type string) is not constant

// TypeCheckError:
// main/files/const26.gno:10:12: v() (value of type string) is not constant
