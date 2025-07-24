package coinflip

import (
	"gno.land/p/moul/txlink"
	"gno.land/p/moul/md"
)

var (
	githubUsername string
	pocInnovationName string
)

func init() {
	githubUsername = "dujardin09"
	pocInnovationName = "PoCInnovation"
}

func renderGame() string {
	out := "# "
	out += md.Link("Heads", txlink.Call("Heads")) + " or "
	out += md.Link("Tails", txlink.Call("Tails")) + " ?\n"

	out += "### Latests Results:\n\n"
	out += GetLatestResults()

	return out
}

func renderIllustration() string {
	return md.Image("Coin Flip Illustration", "https://cdn.dribbble.com/userupload/23852777/file/original-7c6789feec78629e91335795a5baf105.gif")
}

func renderFooter() string {
	out := md.HorizontalRule()
	out += md.BulletList([]string{
		md.Link("Home", "home"),
		md.Link("GitHub: @"+githubUsername, "https://github.com/"+githubUsername),
		md.Link("PoC-Innovation", "https://github.com/"+pocInnovationName),
	})
	return out
}

func Render(_ string) string {
	out := md.Columns([]string{
		renderGame(),
		renderIllustration(),
	}, false)

	out += renderFooter()

	return out
}
