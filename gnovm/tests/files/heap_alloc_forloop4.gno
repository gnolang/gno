package main

type f func()

var fs []f

func forLoopClosure() {
	defer func() {
		for _, f := range fs {
			f()
		}
	}()

	for i := 0; i < 3; i++ {
		fs = append(fs, func() { println(i) })
	}
}

func main() {
	forLoopClosure()
}

// go 1.22 loop var is not supported for now.

// Output:
// 3
// 3
// 3
