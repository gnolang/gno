package main

type fooer interface{}

var f fooer = 1

func trymake() func() {
	switch v := f.(type) {
	case int:
		return func() {
			println(v)
		}
	default:
	}
	return func() {}
}
func main() {
	trymake()()
}

// XXX Fix this by changing Varname to a *NameExpr
// and preprocessing similarly to an expression.

// Output:
// 1
