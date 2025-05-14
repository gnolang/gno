package place

import (
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
	rows uint64 = 45
	cols uint64 = 45

	white string = "rgb(250,250,250)"

	chaeWallet = "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g"
	ghhvWallet = "g1d2hvkqhqckhzpaum09c002fesdw89qrcg9zqws"
	louWallet  = "g14x2crt492f84shxa9s7e5ulk7lnf2p3euz7r9n"
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
	var out string

	out += md.H2(md.Link("Submit", txlink.Call("Submit", "x", xStr, "y", yStr, "colorName", selectedColor)) + " | " +
		md.Link("Delete account", txlink.Call("DeleteUser")) + "\n\n")

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
			tableColor[y][x].initial = "X"
		}
	}

	selectedColor := query.Get("color")

	if selectedColor == "" {
		selectedColor = "W"
	}

	out += md.HorizontalRule() + "\n\n"
	for y, row := range tableColor {
		out += "|"
		for x, cell := range row {
			link := ufmt.Sprintf("/r/eva/place?color=%s&x=%d&y=%d", selectedColor, x, y)
			out += md.Link(cell.initial, link)
		}
		out += "|\n\n"
	}
	out += md.HorizontalRule() + "\n\n"

	out += renderAccountManagement(selectedColor)

	out += ufmt.Sprintf("![Pixel Canvas](data:image/svg+xml;base64,%s)", GenerateSVG())

	for _, color := range colors {
		query.Set("color", color.initial)

		out += ufmt.Sprintf("[!["+color.initial+`](data:image/svg+xml;base64,%s)](`+req.String()+")", generateSquareSvg(color.rgb)) + "\n\n" // I would have liked using md.Link here but doing so replaces the svg image with its link
	}

	return out
}
