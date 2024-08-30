package home

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/art/gnoface"
	"gno.land/r/demo/art/millipede"
	"gno.land/r/leon/config"
)

var (
	pfp string // link to profile picture
)

func init() {
	pfp = "https://i.imgflip.com/91vskx.jpg"
}

func UpdatePFP(url string) {
	config.AssertAuthorized()
	pfp = url
}

func Render(path string) string {
	out := "# Leon's Homepage\n\n"

	out += renderAboutMe()
	out += "\n\n"
	out += renderArt()

	return out
}

func renderAboutMe() string {
	out := "<div class='columns-3'>"

	out += "<div>\n\n"
	out += ufmt.Sprintf("![my profile pic](%s)\n\n[My favourite painting](https://en.wikipedia.org/wiki/Wanderer_above_the_Sea_of_Fog)\n\n", pfp)
	out += "</div>\n\n"

	out += "<div>\n\n"
	out += "I'm Leon, a DevRel engineer at gno.land!\n\n"
	out += "</div>\n\n"

	out += "<div>\n\n"
	out += "I am a tech enthusiast, life-long learner, and sharer of knowledge.\n\n"
	out += "</div>\n\n"

	out += "</div>\n\n"

	return out
}

func renderArt() string {
	out := `<div class="jumbotron">` + "\n\n"
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
