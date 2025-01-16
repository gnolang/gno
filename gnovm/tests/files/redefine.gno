package main

var ss = []int{1, 2, 3}

func main() {
	for _, s := range ss {
		s := s
		println(s)
	}
}

// Output:
// 1
// 2
// 3
