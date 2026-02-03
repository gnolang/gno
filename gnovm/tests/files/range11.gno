package main

func main() {
	i := []int{1, 2, 3}
	for j := range &i {
		println(j)
	}
}

// Error:
// main/range11.gno:5:2-7:3: range iteration over pointer requires array elem type

// TypeCheckError:
// main/range11.gno:5:17: cannot range over &i (value of type *[]int)
