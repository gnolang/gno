package main

import "fmt"

func v() {
	return
}

func main() {
	const t = v()
	fmt.Println(t)
}

// Error:
// main/files/const30.gno:10:8: v<VPBlock(3,0)> (no value) used as value
