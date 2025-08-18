// PKGPATH: gno.land/p/demo/svg
package svg

import "gno.land/p/demo/svg"

func main() {
	canvas := svg.Canvas{
		Width: 500, Height: 500,
		Elems: []svg.Elem{
			svg.Rectangle{X: 50, Y: 50, Width: 100, Height: 100, Fill: "red"},
			svg.Circle{CX: 50, CY: 50, R: 100, Fill: "red"},
			svg.Text{X: 100, Y: 100, Text: "hello world!", Fill: "magenta"},
		},
	}
	println(canvas)
}

// Error:
// gno.land/p/demo/svg/z1.gno:7:12-22: should not happen
