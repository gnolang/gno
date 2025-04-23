package place

import (
	"encoding/base64"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/realmpath"
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
)

func generateSquareSvg(color string) string {
	var svg strings.Builder

	svg.WriteString(`<svg width="80" height="40" xmlns="http://www.w3.org/2000/svg">`)
	svg.WriteString(`<rect x="0" y="0" width="40" height="40" fill="` + color + `"/>`)
	svg.WriteString(`</svg>`)
	return base64.StdEncoding.EncodeToString([]byte(svg.String()))
}

func Render(path string) string {
	req := realmpath.Parse(path)
	query := req.Query
	out := ""

	for _, color := range colors {
		query.Set("color", color.name)
		out += ufmt.Sprintf("[!["+color.name+"](data:image/svg+xml;base64,%s)]("+req.String()+")", generateSquareSvg(color.rgb)) // I would have liked using md.Link here but doing so replaces the svg image with its link
	}
	return out
}
