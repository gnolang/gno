// Package place is a package that reproduces the concept of Reddit's r/place.
// Users can interact with a grid composed of many pixels to color them
// When combined, those pixels can form images
package place

import (
	"net/url"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
)

type Color struct {
	rgb     string
	initial string
}

const (
	rows = 40
	cols = 40

	white = "rgb(250,250,250)"

	// default moderators wallets
	chaeWallet = "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g"
	ghhvWallet = "g1d2hvkqhqckhzpaum09c002fesdw89qrcg9zqws"
	louWallet  = "g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"
)

var (
	colors = []Color{
		{"rgb(0,0,0)", "K"},     // black
		{"rgb(255,0,0)", "R"},   // red
		{"rgb(0,255,0)", "G"},   // green
		{"rgb(0,0,255)", "B"},   // blue
		{"rgb(255,255,0)", "Y"}, // yellow
		{"rgb(0,255,255)", "C"}, // cyan
		{"rgb(255,0,255)", "P"}, // pink
		{white, "W"},            // white
	}

	tableColor [][]Color // TODO: use a B tree for storage efficiency

	xStr string
	yStr string
)

func init() {
	setDefaultAdmin(chaeWallet)
	setDefaultAdmin(ghhvWallet)
	setDefaultAdmin(louWallet)

	tableColor = make([][]Color, rows)
	for y := 0; y < int(rows); y++ {
		tableColor[y] = make([]Color, cols)
		for x := 0; x < int(cols); x++ {
			tableColor[y][x].rgb = white
			tableColor[y][x].initial = "⠀"
		}
	}
}

func renderAccountManagement(selectedColor string) string {
	out := md.H2(md.Link("Submit", txlink.Call("Submit", "x", xStr, "y", yStr, "colorName", selectedColor)) + " | " +
		md.Link("Delete account", txlink.Call("DeleteYourself")) + "\n\n")

	return out
}

func displayCurrentPosition(path string) {
	if strings.Contains(path, "?") {
		x, y, err := parseXY(path)
		if err != nil {
			return
		}
		tableColor[y][x].initial = "X"
	}
}

func displayColorTable(selectedColor string) string {
	out := md.HorizontalRule()
	for y, row := range tableColor {
		out += "|"
		for x, cell := range row {
			link := ufmt.Sprintf("/r/eva/place?color=%s&x=%d&y=%d", selectedColor, x, y)
			out += md.Link(cell.initial, link)
		}
		out += "|\n\n"
	}
	out += md.HorizontalRule()
	return out
}

func displayColorList(query url.Values, req *realmpath.Request) string {
	var out string

	for _, color := range colors {
		query.Set("color", color.initial)

		squareSvg, err := generateSquareSvg(color.rgb)

		if err == nil {
			out += ufmt.Sprintf("[!["+color.initial+`](data:image/svg+xml;base64,%s)](`+req.String()+")", squareSvg) + "\n\n" // I would have liked using md.Link here but doing so replaces the svg image with its link
		}
	}
	return out
}

func Render(path string) string {
	req := realmpath.Parse(path)
	query := req.Query

	out := md.H1("r/place")
	displayCurrentPosition(path)

	selectedColor := query.Get("color")

	if selectedColor == "" {
		selectedColor = "W"
	}

	out += displayColorTable(selectedColor)

	out += renderAccountManagement(selectedColor)

	out += ufmt.Sprintf("![Pixel Canvas](data:image/svg+xml;base64,%s)", generateSVG())

	out += displayColorList(query, req)

	return out
}
