// PKGPATH: gno.land/p/demo/svg_test
package svg_test

import "gno.land/p/demo/svg"

func main() {
	canvas := svg.Canvas{
		Width: 500, Height: 500,
		Elems: []svg.Elem{
			svg.Rectangle{50, 50, 100, 100, "red"},
			svg.Circle{50, 50, 100, "red"},
			svg.Text{100, 100, "hello world!", "magenta"},
		},
	}
	println(canvas)
}

// Output:
// <svg xmlns="http://www.w3.org/2000/svg" width="500" height="500"><rect x="50" y="50" width="100" height="100" fill="red" /><circle cx="50" cy="50" r="100" fill="red" /><text x="100" y="100" fill="magenta">hello world!</text></svg>
