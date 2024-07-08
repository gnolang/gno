/*
Package svg is a minimalist SVG generation library for Gno.

The svg package provides a simple and lightweight solution for programmatically generating SVG (Scalable Vector Graphics) markup in Gno. It allows you to create basic shapes like rectangles and circles, and output the generated SVG to a

Example:

	import "gno.land/p/demo/svg""

	func Foo() string {
	    canvas := svg.Canvas{Width: 200, Height: 200}
	    canvas.DrawRectangle(50, 50, 100, 100, "red")
	    canvas.DrawCircle(100, 100, 50, "blue")
	    return canvas.String()
	}
*/
package svg // import "gno.land/p/demo/svg"
