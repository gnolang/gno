// MAXALLOC: 500000000
package main

func main() {
	var x interface{}
	for i := 0; i < 10000; i++ {
		x = [1]interface{}{x}
	}
}

// Gas:
// 18313861
