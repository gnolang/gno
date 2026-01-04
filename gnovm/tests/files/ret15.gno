package main

func mixedReturns() (a int, _ string, b bool) {
	defer func() {
		b = false
		_ = "blank"
	}()

	a = 5
	b = true
	return a, "ignored", b
}

func main() {
	a, s, b := mixedReturns()
	println(a, s, b)
}

// Output:
// 5 ignored false
