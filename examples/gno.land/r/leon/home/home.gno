package home

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"

	"gno.land/r/demo/art/gnoface"
	"gno.land/r/demo/art/millipede"
	"gno.land/r/demo/mirror"
	"gno.land/r/leon/config"
	"gno.land/r/leon/hor"
)

var (
	pfp        string // link to profile picture
	pfpCaption string // profile picture caption
	abtMe      [2]string
)

func Render(_ string) string {
	out := "# Leon's Homepage\n\n"

	out += renderAboutMe()
	out += renderArt()
	out += config.Banner()

	return out
}

func init() {
	hor.Register("Leon's Home Realm", "")
	mirror.Register(std.CurrentRealm().PkgPath(), Render)

	pfp = "https://i.imgflip.com/91vskx.jpg"
	pfpCaption = "[My favourite painting & pfp](https://en.wikipedia.org/wiki/Wanderer_above_the_Sea_of_Fog)"
	abtMe = [2]string{
		`### About me
Hi, I'm Leon, a DevRel Engineer at gno.land. I am a tech enthusiast, 
life-long learner, and sharer of knowledge.`,
		`### Contributions
My contributions to gno.land can mainly be found 
[here](https://github.com/gnolang/gno/issues?q=sort:updated-desc+author:leohhhn).

TODO import r/gh`,
	}
}

func UpdatePFP(url, caption string) {
	if !config.IsAuthorized(std.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	pfp = url
	pfpCaption = caption
}

func UpdateAboutMe(col1, col2 string) {
	if !config.IsAuthorized(std.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	abtMe[0] = col1
	abtMe[1] = col2
}

func renderAboutMe() string {
	return md.Columns([]string{
		ufmt.Sprintf("![my profile pic](%s)\n\n%s\n", pfp, pfpCaption),
		abtMe[0],
		abtMe[1],
	})
}

func renderArt() string {
	out := "# Gno Art\n"

	out += md.Columns([]string{
		gnoface.Render(strconv.Itoa(int(std.ChainHeight()))),
		renderMillipede(),
		"Empty spot :/",
	})

	out += "This art is dynamic; it will change with every new block.\n\n"

	return out
}

func renderMillipede() string {
	out := "Millipede\n\n"
	out += "```\n" + millipede.Draw(int(std.ChainHeight())%10+1) + "```\n"
	return out
}

func renderBlogPosts() string {
	out := ""
	// out += "## Leon's Blog Posts"

	// todo fetch blog posts authored by @leohhhn
	// and render them
	return out
}
