// MAXALLOC: 500000000
package main

// Each iteration of a three-clause loop whose init declares an array kept
// past its iteration copies that array, and the copy pays per element.

func main() {
	var ps []*[1000]int
	for a := [1000]int{}; a[0] < 3; a[0]++ {
		ps = append(ps, &a)
	}
	println(ps[0][0], ps[2][0])
}

// Output:
// 0 2

// Gas:
// 261675
