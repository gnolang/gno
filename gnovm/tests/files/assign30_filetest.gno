package main

import "fmt"

func t() (int, int) { return 1, 2 }

func main() {
	// v1, v2 := t()
	// fmt.Println("DEFINE call:", v1, v2)	
	
	// var i interface{} = 1
	// iv, ok := i.(int)
	// fmt.Println("DEFINE index:", iv, ok)	

	// m := map[int]string{1: "test"}
	// mv, ok := m[1]
	// fmt.Println("DEFINE map:", mv, ok)	


	// var v1, v2 = t()
	// fmt.Println("VAR call:", v1, v2)	
	
	// var i interface{} = 1
	// var iv, ok = i.(int)
	// fmt.Println("VAR index:", iv, ok)	

	m := map[int]string{1: "test"}
	var mv, ok = m[1]
	fmt.Println("VAR map:", mv, ok)	

	panic("done")

}

// Error:
// test