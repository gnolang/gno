package main

func main() {
	i := [3]int{1, 2, 3}
	for j := range &i {
		println(j)
	}
}

// Output:
// 0
// 1
// 2
