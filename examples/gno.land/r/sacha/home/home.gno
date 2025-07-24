package home

import (
	"gno.land/p/moul/md"
)

var (
	banner string
	title string
	story string
	game [2]string
	art string
	githubUsername string
	pocInnovationName string
)

func init() {
	banner = "https://shorturl.at/yN9Mp"
	story = "In the quiet layers of gno.land, a gnome brushed against a line of forgotten code. No one saw him. No one knows what he changed. But sometimes, in the logs...\n\n\"gnonk.\"\n\nJust once. Then silence."
	title = "Sacha's Space 🏗️"
	game[0] = "🪙 Flip a coin!"
	game[1] = "coinflip"
	art = `⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣾⣿⣿⣷⣦⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣿⣿⣿⣿⣿⣿⣿⣦⡀⠒⢶⣄⠀⠀⠀⠀⠀⠀⠀
⠀⢰⣶⣷⣶⣶⣤⣄⠀⣠⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣾⣿⡆⠀⠀⠀⠀⠀⠀
⠀⢿⣿⣿⣿⣿⡟⢁⣄⠙⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠃⠀⠀⠀⠀⠀⠀
⠀⠘⣿⣿⣿⣿⣧⡈⠻⢷⣦⣄⡉⠛⠿⢿⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠈⠻⣿⣿⣿⣿⣶⣄⡈⠙⠻⢷⣶⣤⣄⣈⡉⠛⠛⠛⠃⢠⣀⣀⡀⠀⠀⠀
⠀⠀⠀⠀⠈⠙⠻⢿⣿⣿⣿⣿⣶⣦⣤⣍⣉⠙⠛⠛⠛⠿⠃⢸⣿⣿⣿⣷⡀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠻⠿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣶⣾⣿⣿⣿⣿⣿⣧⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠉⠙⠛⠻⠏⠀⠉⠻⢿⣿⣿⣿⣿⠿⠋⠀`
	githubUsername = "dujardin09"
	pocInnovationName = "PoCInnovation"
}

func renderStory() string {
	out := md.H2("📖 Story")
	out += md.Paragraph(story)
	return out
}

func renderGame() string {
	out := md.H2("🎰 Game")
	out += md.H3(md.Link(game[0], game[1]))
	return out
}

func renderArt() string {
	out := md.H2("🎨 Art")
	out += md.CodeBlock(art)
	return out
}

func renderBody() string {
	return md.Columns([]string{
		renderStory(),
		renderGame(),
		renderArt(),
	}, false)
}

func renderTitle() string {
	return md.H1(title)
}

func renderBanner() string {
	out := md.Paragraph(md.Image("banner", banner))
	return out
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
	out := renderBanner()
	out += renderTitle()
	out += renderBody()
	out += renderFooter()

	return out
}
