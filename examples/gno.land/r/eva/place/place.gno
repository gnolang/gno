package place

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"gno.land/p/demo/svg"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
)

type Color struct {
	name string
	rgb  string
}

var (
	white = "rgb(240,240,240)"

	colors = []Color{
		{"black", "rgb(40,40,40)"},
		{"red", "rgb(255,0,0)"},
		{"green", "rgb(0,255,0)"},
		{"blue", "rgb(0,0,255)"},
		{"yellow", "rgb(255,255,0)"},
		{"cyan", "rgb(0,255,255)"},
		{"pink", "rgb(255,0,255)"},
		{"white", white},
	}

	table      [][]string // TODO: use a B tree for storage efficiency
	tableColor [][]string
	rows       uint64 = 15 // pour linstant
	cols       uint64 = 80

	xStr string
	yStr string

	ErrInvalidCoordQuery = errors.New("Invalid coordinate value, expecting a uint64")
)

func generateSquareSvg(color string) string {
	// var svg strings.Builder

	canvas := svg.Canvas{Width: 30, Height: 30}
	canvas.DrawRectangle(0, 0, 30, 30, color)

	// svg.WriteString(`<svg width="100" height="100" xmlns="http://www.w3.org/2000/svg">`)
	// svg.WriteString(`<a href="https://fr.wikipedia.org/wiki/Der_Verschwender">`)
	// svg.WriteString(`<rect x="0" y="0" width="20" height="20" fill="` + color + `"/>`)
	// svg.WriteString(`<rect x="30" y="90" width="40" height="10" fill="` + color + `"/>`)
	// svg.WriteString(`<rect x="80" y="0" width="20" height="20" fill="` + color + `"/>`)
	// svg.WriteString(`<text x="0" y="0" fill="` + color + `">Cool dude</text>`)
	// svg.WriteString(`<circle cx="50" cy="50" r="10" fill="` + color + `" />`)
	// svg.WriteString(`<text x="0" y="50" fill="` + color + `">Cool dude</text>`)
	// svg.WriteString(`</a></svg>`)
	return base64.StdEncoding.EncodeToString([]byte(canvas.String()))
	// return svg.String()
}

func init() {

	table = make([][]string, rows)
	for y := 0; y < int(rows); y++ {
		table[y] = make([]string, cols)
		for x := 0; x < int(cols); x++ {
			link := ufmt.Sprintf("/r/eva/place?x=%d&y=%d", x, y)
			table[y][x] = md.Link("⠀", link)
			// table[y][x] = md.Link(generateSquareSvg("rgb(255,255,255)"), link)
		}
	}
	tableColor = make([][]string, rows)
	for y := 0; y < int(rows); y++ {
		tableColor[y] = make([]string, cols)
		for x := 0; x < int(cols); x++ {
			tableColor[y][x] = white
		}
	}
}

// parseXY takes a raw query like "?x=19&y=1" and returns the two ints.
func parseXY(rawQuery string) (x, y uint64, err error) {
	vals, err := url.ParseQuery(strings.TrimPrefix(rawQuery, "?"))
	if err != nil {
		return 0, 0, err
	}

	xStr = vals.Get("x")
	yStr = vals.Get("y")

	x, err = strconv.ParseUint(xStr, 10, 64)
	if err != nil {
		return 0, 0, ErrInvalidCoordQuery
	}
	y, err = strconv.ParseUint(yStr, 10, 64)
	if err != nil {
		return 0, 0, ErrInvalidCoordQuery
	}
	return x, y, nil
}

func Submit(x, y uint64, color string) {
	table[y][x] = "O"
	tableColor[y][x] = color
}

func GenerateSVG() string {
	out := ufmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, cols*40, rows*40)
	for y := 0; y < int(rows); y++ {
		for x := 0; x < int(cols); x++ {
			out += ufmt.Sprintf(`<rect x="%d" y="%d" width="40" height="40" fill="%s"/>`, x*40, y*40, tableColor[y][x])
		}
	}
	out += "</svg>"
	return out
}

func Render(path string) string {
	req := realmpath.Parse(path)
	query := req.Query

	out := md.H1("r/place") + "\n\n"
	if strings.Contains(path, "?") {
		if x, y, err := parseXY(path); err != nil {
			panic(err)
		} else {
			table[y][x] = "X"
		}
	}

	out += md.HorizontalRule() + "\n\n"
	for _, row := range table {
		out += "|"
		for _, cell := range row {
			out += cell
		}
		out += "|\n\n"
	}
	out += md.HorizontalRule() + "\n\n"

	selectedColor := query.Get("color")

	if selectedColor == "" {
		selectedColor = white
	}

	out += "## " + md.Link("Submit", txlink.Call("Submit", "x", xStr, "y", yStr, "color", selectedColor)) + "\n\n"

	ssvg := GenerateSVG()
	base64SVG := base64.StdEncoding.EncodeToString([]byte(ssvg))
	out += ufmt.Sprintf("![Pixel Canvas](data:image/svg+xml;base64,%s)", base64SVG)

	for _, color := range colors {
		query.Set("color", color.name)

		out += ufmt.Sprintf("[!["+color.name+`](data:image/svg+xml;base64,%s)](`+req.String()+")", generateSquareSvg(color.rgb)) // I would have liked using md.Link here but doing so replaces the svg image with its link
		// out += ufmt.Sprintf("!["+color.name+"](data:image/svg+xml;base64,%s)", generateSquareSvg(color.rgb))
	}
	// canvas := svg.Canvas{Width: 200, Height: 200}
	// canvas.DrawRectangle(0, 0, 100, 100, "rgb(0,0,255)")
	// canvas.DrawRectangle(100, 0, 100, 100, "rgb(255,0,255)")
	// out += md.Image("name", "data:image/svg+xml;base64,"+base64.StdEncoding.EncodeToString([]byte(canvas.String())))
	// canvas = svg.Canvas{Width: 200, Height: 200}
	// // canvas.DrawRectangle(0, 0, 100, 100, "rgb(0,0,255)")
	// // canvas.DrawRectangle(100, 0, 100, 100, "rgb(255,0,255)")
	// canvas.DrawText(0, 0, "Cool dude", "rgb(0,0,255)")
	// out += md.Image("name", "data:image/svg+xml;base64,"+base64.StdEncoding.EncodeToString([]byte(canvas.String())))

	// ufmt.Printf("ICI%sICI", canvas.String())
	return out
}
