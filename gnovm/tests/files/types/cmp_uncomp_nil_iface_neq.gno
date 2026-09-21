package main

// One unassigned interface{} and one carrying a value compare unequal.
// Exercises isEql's one-sided IsUndefined() early-return path.
func main() {
	var x interface{}
	var y interface{} = 5
	println(x == y)
}

// Output:
// false
