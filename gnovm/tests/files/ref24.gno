package main

import "fmt"

// Auto-address for pointer receiver via selector expression.
// When t.PointerMethod() is called and t is addressable, the
// preprocessor synthesizes &t, setting ATTR_TYPEOF_VALUE on t
// before wrapping it in RefExpr.  Exercises the setPreprocessed
// path in SelectorExpr TRANS_LEAVE.

type Counter struct{ n int }

func (c *Counter) Inc() { c.n++ }

func main() {
	var c Counter
	c.Inc()
	c.Inc()
	fmt.Println(c.n)
}

// Output:
// 2
