package main

func main() {
	_ = &func() { println("Hello, World!") }
}

// Error:
// main/files/addressable_8a_err.gno:4:6: cannot take address of func func(){ (const (println func(xs ...interface{})()))((const ("Hello, World!" string))) }
