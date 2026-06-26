package main

func f(xs []int) {}

func main() {
	var v func(...int) = f
	_ = v
}

// Error:
// main/assign_func_varg.gno:6:6-24: cannot use func([]int) as func(...int)

// TypeCheckError:
// main/assign_func_varg.gno:6:23: cannot use f (value of type func(xs []int)) as func(...int) value in variable declaration
