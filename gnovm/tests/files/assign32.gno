package main

func foo() int {
	return 2
}

func main() {
	var mp map[string]int = map[string]int{"idx": 4}
	var sl []int = []int{4, 5, 6}
	arr := [1]int{7}
	var num any = 5

	a, b, c, d, e, f, g := int(1), foo(), 3, mp["idx"], num.(int), sl[2], arr[0]
	println(a, b, c, d, e, f, g)

	var h, i, j, k, l, m, n int = int(1), foo(), 3, mp["idx"], num.(int), sl[2], arr[0]
	println(h, i, j, k, l, m, n)
}

// Output:
// 1 2 3 4 5 6 7
// 1 2 3 4 5 6 7
