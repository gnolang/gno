package main

var m1 = make(map[int]string)

var m2 = make(map[int]int)

func init() {
	m2[1] = 1
	m1[m2[1]] = "a"
}
func main() {
	println(m1[m2[1]])
}

// Output
