/*
Package svg is a minimalist and extensible SVG generation library for Gno.

It provides a structured way to create and compose SVG elements such as rectangles, circles, text, paths, and more. The package is designed to be modular and developer-friendly, enabling optional attributes and method chaining for ease of use.

Each SVG element embeds a BaseAttrs struct, which supports common SVG attributes like `id`, `class`, `style`, `fill`, `stroke`, and `transform`.

Canvas objects represent the root SVG container and support global dimensions, viewBox configuration, embedded styles, and element composition.

Example:

	import "gno.land/p/demo/svg"

	func Foo() string {
		canvas := svg.NewCanvas(200, 200).WithViewBox(0, 0, 200, 200)
		canvas.AddStyle(".my-rect", "stroke:black;stroke-width:2")
		canvas.Append(
			svg.NewRectangle(60, 40, 100, 50, "red").WithClass("my-rect"),
			svg.NewCircle(50, 80, 40, "blue"),
			&svg.Path{D: `M 10,30
			A 20,20 0,0,1 50,30
				A 20,20 0,0,1  90,30
				Q 90,60 50,90
				Q 10,60 10,30 z`, Fill: "magenta"},
			svg.NewText(20, 50, "Hello SVG", "black"),
		)
		mysvg := canvas.Base64()
	}
*/
package svg // import "gno.land/p/demo/svg"
