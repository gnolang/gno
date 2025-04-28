package place

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
)

var (
	table [][]string
	rows  uint64 = 15
	cols  uint64 = 91

	xStr string
	yStr string

	ErrInvalidCoordQuery = errors.New("Invalid coordinate value, expecting a uint64")
)

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

func init() {
	// rows, cols := 15, 91

	table = make([][]string, rows)
	for y := 0; y < int(rows); y++ {
		table[y] = make([]string, cols)
		for x := 0; x < int(cols); x++ {
			link := ufmt.Sprintf("/r/eva/place?x=%d&y=%d", x, y)
			table[y][x] = md.Link("⠀", link)
		}
	}
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
	return out
}
