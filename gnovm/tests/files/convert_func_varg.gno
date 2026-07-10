package main

func f(xs []int) {}

func main() {
	_ = (func(...int))(f)
}

// Error:
// main/convert_func_varg.gno:6:6-23: cannot convert f<VPBlock(3,0)> (of type func([]int)) to type func(...int)

// TypeCheckError:
// main/convert_func_varg.gno:6:21: cannot convert f (value of type func(xs []int)) to type func(...int)
