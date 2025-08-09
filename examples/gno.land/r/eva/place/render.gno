package place

import (
	"net/url"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
)

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

// Helpers
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
			out += ufmt.Sprintf("[!["+color.initial+`](data:image/svg+xml;base64,%s)](`+req.String()+")", squareSvg) + " " // I would have liked using md.Link here but doing so replaces the svg image with its link
		}
	}
	return out
}
