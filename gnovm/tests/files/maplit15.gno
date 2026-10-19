// Map composite-literal keys referencing an enclosing for-loop variable
// (#5910). Origin: test/map.go from Go's test corpus,
// https://github.com/golang/go/blob/master/test/map.go.
package main

type M map[int]int

func main() {
	for i := 0; i < 3; i++ {
		m := map[int]int{i: i + 1}
		n := M{i: i + 1}
		println(m[i], n[i])
	}
}

// Output:
// 1 1
// 2 2
// 3 3
