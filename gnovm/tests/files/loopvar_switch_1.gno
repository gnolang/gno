package main

// switch on loop var: the loop var expression is used in the switch tag.
// Closures capture a body-scope label derived from the switch result.
func main() {
	var fns []func() string
	for i := 0; i < 4; i++ {
		var label string
		switch i {
		case 0:
			label = "zero"
		case 1:
			label = "one"
		default:
			label = "other"
		}
		fns = append(fns, func() string { return label })
	}
	for _, fn := range fns {
		println(fn())
	}
}

// Output:
// zero
// one
// other
// other
