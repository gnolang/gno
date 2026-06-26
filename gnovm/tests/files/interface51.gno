package main

type I interface {
	M(...int)
}

type T struct{}

func (T) M([]int) {}

func main() {
	var i I = T{}
	_ = i
}

// Error:
// main/interface51.gno:12:6-15: main.T does not implement main.I (wrong type for method M)

// TypeCheckError:
// main/interface51.gno:12:12: cannot use T{} (value of struct type T) as I value in variable declaration: T does not implement I (wrong type for method M)
// 		have M([]int)
// 		want M(...int)
