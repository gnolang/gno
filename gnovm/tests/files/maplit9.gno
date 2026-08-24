package main

func main() {
	s1 := []int{1, 2}
	s2 := []int{2, 3}
	_ = map[*[]int]int{&s1: 10, &s2: 11}
	println("ok")
}

// Output:
// ok
