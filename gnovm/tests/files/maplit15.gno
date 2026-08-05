package main

func main() {
	for i := 0; i < 3; i++ {
		m := map[int]int{i: i + 1}
		println(m[i])
	}
}

// Output:
// 1
// 2
// 3
