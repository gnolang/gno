package main

const a = any(nil)

func main() {
	println("ok")
}

// Error:
// main/const44.gno:3:7-19: (const (undefined)) (variable of type interface {}) is not constant

// TypeCheckError:
// main/const44.gno:3:11: any(nil) (value of interface type any) is not constant
