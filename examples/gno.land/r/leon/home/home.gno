package home

import (
	"chain/banker"
	"chain/runtime"
	"strconv"

	"gno.land/p/demo/svg"
	"gno.land/p/leon/svgbtn"
	"gno.land/p/moul/md"
	"gno.land/p/nt/ufmt"

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

func Render(path string) string {
	out := "# Leon's Homepage\n\n"

	if path == "buttons" {
		return renderButtonPage()
	}

	out += renderAboutMe()
	out += renderArt()
	out += config.Banner()
	out += "\n\n"
	out += svgbtn.Button(
		1200,
		50,
		gnomeBodyColors[int(runtime.ChainHeight()+1)%len(gnomeBodyColors)],
		"#ffffff",
		"Support my work!",
		"/r/leon/home$help&func=Donate&.send=1000000ugnot",
	)

	return out
}

func init() {
	hor.Register(cross, "Leon's Home Realm", "")
	mirror.Register(cross, runtime.CurrentRealm().PkgPath(), Render)

	pfp = "https://i.imgflip.com/91vskx.jpg"
	pfpCaption = "[My favourite painting & pfp](https://en.wikipedia.org/wiki/Wanderer_above_the_Sea_of_Fog)"
	abtMe = [2]string{
		`### About me
Hi, I'm Leon, a DevRel Engineer at gno.land. I am a tech enthusiast, 
life-long learner, and sharer of knowledge.`,
		`### Contributions
My contributions to Gno.land can mainly be found 
[on GitHub](https://github.com/gnolang/gno/issues?q=sort:updated-desc+author:leohhhn), and on [the chain](/u/leon).

TODO import r/gh`,
	}
}

func UpdatePFP(cur realm, url, caption string) {
	if !config.IsAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	pfp = url
	pfpCaption = caption
}

func UpdateAboutMe(cur realm, col1, col2 string) {
	if !config.IsAuthorized(runtime.PreviousRealm().Address()) {
		panic(config.ErrUnauthorized)
	}

	abtMe[0] = col1
	abtMe[1] = col2
}

func Donate(_ realm) string {
	sent := banker.OriginSend()
	if len(sent) == 0 {
		return ":c"
	}

	banker.NewBanker(banker.BankerTypeOriginSend).SendCoins(
		runtime.CurrentRealm().Address(),
		config.OwnableMain.Owner(),
		sent,
	) // wish this was prettier :)

	return "Thanks for donating " + sent.String() + " <3"
}

func renderAboutMe() string {
	return md.Columns([]string{
		ufmt.Sprintf("![my profile pic](%s)\n\n%s\n", pfp, pfpCaption),
		abtMe[0],
		abtMe[1],
	}, false)
}

func renderArt() string {
	out := "# Gno Art\n"

	out += md.Columns([]string{
		gnoface.Render(strconv.Itoa(int(runtime.ChainHeight()))),
		renderMillipede(),
		"SVG Gnome\n" + RenderSVGGnome(),
	}, false)

	out += "This art is dynamic; it will change with every new block.\n\n"

	return out
}

func renderMillipede() string {
	out := "Millipede\n\n"
	out += "```\n" + millipede.Draw(int(runtime.ChainHeight())%10+1) + "```\n"
	return out
}

func renderBlogPosts() string {
	out := ""
	// out += "## Leon's Blog Posts"

	// todo fetch blog posts authored by @leohhhn
	// and render them
	return out
}

func RenderSVGGnome() string { // exported for your pleasure :)
	c := svg.NewCanvas(430, 430).WithViewBox(13, 25, 75, 75)

	// Body: blue triangle
	body := svg.NewPolygon("50,50 30,100 70,100", gnomeBodyColors[int(runtime.ChainHeight())%len(gnomeBodyColors)])

	// Head: peach circle (overlaps body)
	head := svg.NewCircle(50, 60, 10, "#FAD7B6")

	// Hat: red triangle on top of head
	hat := svg.NewPolygon("50,30 35,55 65,55", "#E53935")

	// Eyes: two small black dots
	leftEye := svg.NewCircle(46, 59, 1, "#000")
	rightEye := svg.NewCircle(54, 59, 1, "#000")

	// Beard: small white triangle under head
	beard := svg.NewPolygon("50,85 42,63 58,63", "#FFF")

	// Layering order matters (bottom to top)
	c.Append(body, head, beard, hat, leftEye, rightEye)

	return c.Render("svg gnome")
}

func renderButtonPage() string {
	out := "# Buttons Demo\n\n"

	out += md.ColumnsN([]string{
		svgbtn.PrimaryButton(140, 45, "Click Me", "/r/leon/home:click") + "\n\n",
		svgbtn.DangerButton(140, 45, "Delete", "/delete") + "\n\n",
		svgbtn.SuccessButton(140, 45, "Go Home", "/r/leon/home") + "\n\n",
		svgbtn.SmallButton(100, 45, "Edit", "/edit") + "\n\n",
		svgbtn.WideButton(200, 40, "Big Action", "/big") + "\n\n",
		svgbtn.TextButton(100, 30, "More Info", "/r/leon/home:info") + "\n\n",
		svgbtn.IconButton(100, 40, "Config", "/r/leon/config") + "\n\n",
	}, 3, true)

	return out
}

var gnomeBodyColors = []string{
	"#4CAF50", // Green
	"#2196F3", // Blue
	"#9C27B0", // Purple
	"#FF5722", // Orange
	"#795548", // Brown
	"#607D8B", // Grayish Blue
	"#E91E63", // Pink
	"#FFC107", // Amber
	"#00BCD4", // Cyan
	"#8BC34A", // Light Green
	"#FF9800", // Deep Orange
	"#3F51B5", // Indigo
	"#673AB7", // Deep Purple
	"#009688", // Teal
	"#F44336", // Red
	"#CDDC39", // Lime
}
