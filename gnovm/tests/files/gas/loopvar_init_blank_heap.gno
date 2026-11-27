// MAXALLOC: 500000000
package main

// A three-clause loop whose init declares "_" and whose body keeps a pointer
// to a [1000]int. Only the init's named variable is copied per iteration: the
// body's array, which the next iteration declares afresh, is not.

func two() (int, int) { return 0, 0 }

func main() {
	var ps []*[1000]int
	for _, x := two(); x < 10; x++ {
		big := [1000]int{}
		ps = append(ps, &big)
	}
	println(len(ps))
}

// Output:
// 10

// Gas:
// 1002357
