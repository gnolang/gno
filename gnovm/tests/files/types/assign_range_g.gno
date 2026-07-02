package main

const c = 5

func main() {
	a := []int{1, 2, 3}
	for _, c = range a {
	}
	println("ok")
}

// Error:
// main/assign_range_g.gno:7:2-8:3: cannot assign to (const (5 <untyped> bigint))

// TypeCheckError:
// main/assign_range_g.gno:7:9: cannot assign to c (neither addressable nor a map index expression)
