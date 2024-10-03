package main

func main() {
	const x = 118999
	y := 11
	p := []int{}
	p = append(p, x, y)
	println(p[0] + p[1])
}

// Output:
// 119010
