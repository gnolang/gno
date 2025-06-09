// PKGPATH: gno.land/p/demo/svg_test
package svg_test

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

// Output:
// <svg xmlns="http://www.w3.org/2000/svg" width="500" height="500" viewBox=""><rect x="50" y="50" width="100" height="100" rx="0" ry="0" fill="red" /><circle cx="100" cy="100" r="50" fill="blue" /><text x="100" y="100" dx="0" dy="0" rotate="" fill="magenta" >hello world!</text><circle cx="100" cy="100" r="50" fill="blue" class="toto"/></svg>
