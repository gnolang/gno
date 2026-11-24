package main

import "strings"

// A closure inside the string argument of append or copy, in a plain block
// inside another closure, must still capture the outer variable.
func render(parts []string) string {
	sep := ", "
	out := func() string {
		var b []byte
		for i, p := range parts {
			if i > 0 {
				b = append(b, func() string { return strings.ToUpper(sep) }()...)
			}
			b = append(b, p...)
		}
		c := make([]byte, 2)
		{
			copy(c, func() string { return sep }())
		}
		return string(b) + "|" + string(c)
	}
	return out()
}

func main() {
	println(render([]string{"a", "b"}))
}

// Output:
// a, b|,
