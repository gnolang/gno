package main

// goto label inside for loop body, with loop var references on both sides of the label.
func main() {
	for i := 0; i < 3; i++ {
		if i == 1 {
			goto SKIP
		}
		println("body:", i)
	SKIP:
		println("end:", i)
	}
}

// Output:
// body: 0
// end: 0
// end: 1
// body: 2
// end: 2
