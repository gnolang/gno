package home

import (
	"std"
	"strconv"

	"gno.land/r/demo/art/gnoface"
	"gno.land/r/demo/art/millipede"
)

func Render(path string) string {
	out := ""

	out += "# Leon's Homepage\n\n"
	out += `<div class="jumbotron">` + "\n\n"
	out += "# Gno Art\n\n"

	out += "<div class='columns-3'>"

	out += renderGnoFace()
	out += renderMillipede()
	out += "Empty spot :/"

	out += "</div>\n\n"
	out += "This art is dynamic; it will change with every new block.\n\n"
	out += `</div><!-- /jumbotron -->` + "\n"

	return out
}

func renderGnoFace() string {
	out := "<div>\n\n"
	out += gnoface.Render(strconv.Itoa(int(std.GetHeight())))
	out += "</div>\n\n"

	return out
}

func renderMillipede() string {
	out := "<div>\n\n"
	out += "Millipede\n\n"
	out += "```\n" + millipede.Draw(int(std.GetHeight())%10+1) + "```\n"
	out += "</div>\n\n"

	return out
}
