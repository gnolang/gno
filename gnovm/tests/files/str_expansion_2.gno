// MAXALLOC: 100000000
package main

func main() {
	var hugeString string
	hugeString += string(make([]byte, 9000000))
	hugeString += string(make([]byte, 20000000))
	// or  hugeString = hugeString + string(make([]byte, 20000000))
	println("ok")
}

// Output:
// ok
