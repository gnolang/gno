package main

func main() {
	v := T
	i := 0
	for v {
		if i > 2 {
			break
		}
		println(i)
		i++
	}
}

type C bool

const (
	F C = false
	T C = true
)

// Output:
// 0
// 1
// 2
