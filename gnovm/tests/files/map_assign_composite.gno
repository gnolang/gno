package main

// Assigning an array or struct to a new map key stores the value given, and
// op-assignments on a missing key still start from the zero value. Ranging
// over the map with "_" for the key or the value binds nothing to it.

type S struct {
	A [2]int
	B string
}

func main() {
	m := map[int]S{}
	m[1] = S{B: "x"}
	m[2] = S{A: [2]int{3, 4}}
	println(m[1].A[1], m[1].B, m[2].A[1], len(m))

	n := map[string][2]int{}
	n["k"] = [2]int{5, 6}
	println(n["k"][1], n["missing"][0], len(n))

	c := map[string]int{}
	c["a"] += 2
	c["b"]++
	println(c["a"], c["b"])

	sum := 0
	for k, _ := range m {
		sum += k
	}
	for _, v := range n {
		sum += v[0]
	}
	var k2 string
	for k2, _ = range n {
	}
	println(sum, k2)
}

// Output:
// 0 x 4 2
// 6 0 1
// 2 1
// 8 k
