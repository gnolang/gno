package home

import (
	"std"

	"gno.land/p/demo/svg"
	"gno.land/p/lou/tokenaccess"
	"gno.land/p/moul/md"
)

func renderSVG() string {
	canvas := svg.Canvas{
		Width:  420,
		Height: 340,
	}
	canvas.AddStyle(".cls-1, .cls-2, .cls-3, .cls-4", "stroke-width: 0px;")
	canvas.AddStyle(".cls-2", "fill: #f9d889;")
	canvas.AddStyle(".cls-3", "fill: #f9fafc;")
	canvas.AddStyle(".cls-4", "fill: #124699;")

	canvas.Append(
		svg.Path{
			D: string(svgAssets["strokes"]),
			Attr: svg.BaseAttrs{
				Class: "cls-1",
			},
		},
		svg.Path{
			D: string(svgAssets["blues"]),
			Attr: svg.BaseAttrs{
				Class: "cls-4",
			},
		},
		svg.Path{
			D: string(svgAssets["yellows"]),
			Attr: svg.BaseAttrs{
				Class: "cls-2",
			},
		},
		svg.Path{
			D: string(svgAssets["whites"]),
			Attr: svg.BaseAttrs{
				Class: "cls-3",
			},
		},
	)
	return "![svg](data:image/svg+xml;base64," + canvas.Base64() + ")"
}

func Render(_ string) string {
	out := "<gno-columns>\n\n"
	out += md.Image("Github profile picture", "https://avatars.githubusercontent.com/u/109749395?v=4&width=30") + "\n\n"
	out += "|||\n\n"

	out += "# Lou's Home\n"
	out += md.HorizontalRule()
	out += RenderRockPaperScissors()
	if err := tokenaccess.RequireMinimumAllowance(token, admin, 5, std.Address(currentProfile.Account())); err != nil {
		out += md.H2("No further content unlocked.") + "\n\n"
		out += md.Italic("--> "+err.Error()) + "\n\n"
	} else {
		out += renderMagic8Ball()
	}
	out += "</gno-columns>\n\n"

	out += renderSVG() + "\n\n"

	out += renderRatings()
	return out
}
