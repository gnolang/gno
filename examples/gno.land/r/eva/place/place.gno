package place

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strconv"
	"strings"

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
	colors = []Color{
		{"red", "rgb(255,0,0)"},
		{"green", "rgb(0,255,0)"},
		{"blue", "rgb(0,0,255)"},
	}

	table [][]string
	rows  uint64 = 15 // pour linstant
	cols  uint64 = 91

	xStr string
	yStr string

	ErrInvalidCoordQuery = errors.New("Invalid coordinate value, expecting a uint64")
)

func generateSquareSvg(color string) string {
	var svg strings.Builder

	svg.WriteString(`<span><svg viewBox="0 0 100 100" width="10" height="10" xmlns="http://www.w3.org/2000/svg">`)
	svg.WriteString(`<rect x="0" y="0" width="10" height="10" fill="` + color + `"/>`)
	svg.WriteString(`</svg></span>`)
	return base64.StdEncoding.EncodeToString([]byte(svg.String()))
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

func Submit(x, y uint64) {
	table[y][x] = "**O**"
}

func GenerateSVG() string {
	out := `<svg width="3600" height="600" xmlns="http://www.w3.org/2000/svg">`
	out += `<rect x="0" y="0" width="3600" height="600" fill="rgb(20,20,20)"/>`
	for y := 0; y < int(rows); y++ {
		for x := 0; x < int(cols); x++ {
			if table[y][x] == "**O**" || table[y][x] == "**X**" {
				out += ufmt.Sprintf(`<rect x="%d" y="%d" width="40" height="40" fill="rgb(255, 255, 255)"/>`, x*40, y*40)
			}
		}
	}
	out += "</svg>"
	return out
}

func Render(path string) string {
	out := md.H1("r/place") + "\n\n"
	if strings.Contains(path, "?") {
		if x, y, err := parseXY(path); err != nil {
			panic(err)
		} else {
			table[y][x] = "**X**"
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

	out += "## " + md.Link("Submit", txlink.Call("Submit", "x", xStr, "y", yStr)) + "\n\n"

	svg := GenerateSVG()
	base64SVG := base64.StdEncoding.EncodeToString([]byte(svg))
	out += ufmt.Sprintf("![Pixel Canvas](data:image/svg+xml;base64,%s)", base64SVG)

	req := realmpath.Parse(path)
	query := req.Query
	out += ""

	for _, color := range colors {
		query.Set("color", color.name)
		out += ufmt.Sprintf("[!["+color.name+"](data:image/svg+xml;base64,%s)]("+req.String()+")", generateSquareSvg(color.rgb)) + " help" // I would have liked using md.Link here but doing so replaces the svg image with its link
	}

	return out
}
