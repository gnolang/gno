// Range over a pointer to array iterates the array, so the value target's
// type must be assignable from the array element type. The pointer must be
// unwrapped to the array type, else this check is silently skipped.
package main

func main() {
	a := [3]int{1, 2, 3}
	var v string
	for _, v = range &a {
	}
	_ = v
	println("ok")
}

// Error:
// main/assign_range_l.gno:9:2-10:3: cannot use int as string

// TypeCheckError:
// main/assign_range_l.gno:9:9: cannot use v (value of type int) as string value in assignment
