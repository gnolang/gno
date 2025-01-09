package main

import "fmt"

func main() {
	var s []string = []string{"1"}
	const t, ok = s[0]
	fmt.Println(t, ok)
}

// Error:
// main/files/const28.gno:7:8: s<VPBlock(1,0)>[(const (0 int))] (variable of type string) is not constant
