package main

const a = interface{}(nil)

func main() {
	println("ok")
}

// Error:
// main/files/const44.gno:3:7: (const (undefined)) (variable of type interface{}) is not constant
