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
	pfp        string // link to profile picture
	pfpCaption string // profile picture caption
	abtMe      [2]string
)

func init() {
	pfp = "https://i.imgflip.com/91vskx.jpg"
	pfpCaption = "[My favourite painting](https://en.wikipedia.org/wiki/Wanderer_above_the_Sea_of_Fog)"
	abtMe = [2]string{
		"Hi, I'm Leon, a DevRel Engineer at gno.land.",
		"I am a tech enthusiast, life-long learner, and sharer of knowledge.",
	}
}

func UpdatePFP(url, caption string) {
	config.AssertAuthorized()
	pfp = url
	pfpCaption = caption
}

func UpdateAboutMe(col1, col2 string) {
	config.AssertAuthorized()
	abtMe[0] = col1
	abtMe[1] = col2
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
	out += ufmt.Sprintf("![my profile pic](%s)\n\n%s\n\n", pfp, pfpCaption)
	out += "</div>\n\n"

	out += "<div>\n\n"
	out += abtMe[0] + "\n\n"
	out += "</div>\n\n"

	out += "<div>\n\n"
	out += abtMe[1] + "\n\n"
	out += "\n\n"
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
