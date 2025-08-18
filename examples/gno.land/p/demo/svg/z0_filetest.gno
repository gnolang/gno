// PKGPATH: gno.land/p/demo/svg
package svg

import "gno.land/p/demo/svg"

func main() {
	canvas := svg.Canvas{Width: 500, Height: 500}
	canvas.Append(
		svg.Rectangle{X: 50, Y: 50, Width: 100, Height: 100, Fill: "red"},
		svg.Circle{CX: 100, CY: 100, R: 50, Fill: "blue"},
		svg.Text{X: 100, Y: 100, Text: "hello world!", Fill: "magenta"},
	)
	canvas.Append(
		svg.NewCircle(100, 100, 50, "blue").WithClass("toto"),
	)
	println(canvas)
}

// Error:
// gno.land/p/demo/svg/z0.gno:7:12-22: should not happen
