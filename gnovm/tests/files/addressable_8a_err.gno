package main

func main() {
	_ = &func() { println("Hello, World!") }
}

// Error:
// main/addressable_8a_err.gno:4:6-42: invalid operation: cannot take address of func func(){ (const (println func(...interface {})))((const ("Hello, World!" string))) } (value of type func())

// TypeCheckError:
// main/addressable_8a_err.gno:4:7: invalid operation: cannot take address of (func() literal) (value of type func())
