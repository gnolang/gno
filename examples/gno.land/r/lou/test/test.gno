package svg

import (
	"gno.land/p/demo/svg"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
)

func Render(path string) string {
	out := md.H2("SVG Example!!")
	// canvas := svg.NewCanvas(200, 200).WithViewBox(0, 0, 200, 200)
	// canvas.AddStyle(".my-rect", "stroke:black;stroke-width:2")
	// canvas.Append(
	// 	svg.NewRectangle(60, 40, 100, 50, "red").WithClass("my-rect"),
	// 	svg.NewCircle(50, 80, 40, "blue"),
	// 	svg.Path{D: `M 10,30
	// 	A 20,20 0,0,1 50,30
	// 	A 20,20 0,0,1 90,30
	// 	Q 90,60 50,90
	// 	Q 10,60 10,30 z`, Fill: "magenta"},
	// 	svg.NewText(20, 50, "Hello SVG", "black"),
	// )
	// canvas := svg.Canvas{Width: 500, Height: 500}
	// canvas.Append(
	// 	svg.Rectangle{X: 50, Y: 50, Width: 100, Height: 100, Fill: "red"},
	// 	svg.Circle{CX: 100, CY: 100, R: 50, Fill: "blue"},
	// 	svg.Text{X: 100, Y: 100, Text: "hello world!", Fill: "magenta"},
	// )
	// canvas.Append(
	// 	svg.NewCircle(100, 100, 50, "blue").WithClass("toto"),
	// )
	canvas := svg.Canvas{
		Width: 500, Height: 500,
		Elems: []svg.Elem{
			svg.Rectangle{X: 50, Y: 50, Width: 100, Height: 100, Fill: "red"},
			svg.Circle{CX: 50, CY: 50, R: 100, Fill: "red"},
			svg.Text{X: 100, Y: 100, Text: "hello world!", Fill: "magenta"},
		},
	}
	mysvg := canvas.Base64()

	// canvas.Append(svg.Rectangle{X: 50, Y: 50, Width: 100, Height: 100, Fill: "red"})
	// canvas.Append(svg.Circle{CX: 100, CY: 100, R: 50, Fill: "yellow"})
	// canvas.Append(svg.Text{X: 100, Y: 100, Text: "hello world!", Fill: "magenta"})

	// out += "\n\n" + ufmt.Sprintf("[%s]\n\n", canvas.String())
	out += "\n\n" + ufmt.Sprintf("![heyy](data:image/svg+xml;base64,%s)\n\n[%s]\n", mysvg, mysvg)

	return out
}
