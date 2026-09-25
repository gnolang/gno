package main

type T struct{ x int }

func (t T) M(a, b int) { _ = t.x + a + b }

func main() {
	var t T
	for i := 0; i < 10; i++ {
		defer t.M(1, 2)
	}
}

// Gas:
// 42420
