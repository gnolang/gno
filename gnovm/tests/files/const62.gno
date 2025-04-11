package main

import "fmt"

func main() {
	const a = len([1]int{
		func() int { println("hey!!"); return 1 }(),
	})
	fmt.Println("Hello", a)
}

// Error:
// main/files/const62.gno:6:8: (const (len func(interface {}) int))([(const (1 int))](const-type int){func func()  (const-type int){ (const (println func(...interface {})))((const ("hey!!" string))); return (const (1 int)) }()}) (value of type [1]int) is not constant
