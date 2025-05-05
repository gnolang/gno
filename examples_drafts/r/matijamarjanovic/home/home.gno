package home

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/r/leon/hor"
)

var (
	pfp        string // link to profile picture
	pfpCaption string // profile picture caption
	abtMe      string

	modernVotes  int64
	classicVotes int64
	minimalVotes int64
	currentTheme string

	modernLink  string
	classicLink string
	minimalLink string
)

func init() {
	pfp = "https://static.artzone.ai/media/38734/conversions/IPF9dR7ro7n05CmMLLrXIojycr1qdLFxgutaaanG-w768.webp"
	pfpCaption = "My profile picture - Tarantula Nebula"
	abtMe = `Motivated Computer Science student with strong
                analytical and problem-solving skills. Proficient in
                programming and version control, with a high level of
                focus and attention to detail. Eager to apply academic
                knowledge to real-world projects and contribute to
                innovative technology solutions.
                In addition to my academic pursuits,
                I enjoy traveling and staying active through weightlifting.
                I have a keen interest in electronic music and often explore various genres.
                I believe in maintaining a balanced lifestyle that complements my professional development.`

	modernVotes = 0
	classicVotes = 0
	minimalVotes = 0
	currentTheme = "classic"
	modernLink = "https://www.google.com"
	classicLink = "https://www.google.com"
	minimalLink = "https://www.google.com"
	hor.Register("Matija Marijanovic's Home Realm", "")
}

func UpdatePFP(url, caption string) {
	AssertAuthorized()
	pfp = url
	pfpCaption = caption
}

func UpdateAboutMe(col1 string) {
	AssertAuthorized()
	abtMe = col1
}

func maxOfThree(a, b, c int64) int64 {
	max := a
	if b > max {
		max = b
	}
	if c > max {
		max = c
	}
	return max
}

func VoteModern() {
	ugnotAmount := std.OriginSend().AmountOf("ugnot")
	votes := ugnotAmount
	modernVotes += votes
	updateCurrentTheme()
}

func VoteClassic() {
	ugnotAmount := std.OriginSend().AmountOf("ugnot")
	votes := ugnotAmount
	classicVotes += votes
	updateCurrentTheme()
}

func VoteMinimal() {
	ugnotAmount := std.OriginSend().AmountOf("ugnot")
	votes := ugnotAmount
	minimalVotes += votes
	updateCurrentTheme()
}

func updateCurrentTheme() {
	maxVotes := maxOfThree(modernVotes, classicVotes, minimalVotes)

	if maxVotes == modernVotes {
		currentTheme = "modern"
	} else if maxVotes == classicVotes {
		currentTheme = "classic"
	} else {
		currentTheme = "minimal"
	}
}

func CollectBalance() {
	AssertAuthorized()

	banker := std.NewBanker(std.BankerTypeRealmSend)
	ownerAddr := Address()

	banker.SendCoins(std.CurrentRealm().Address(), ownerAddr, banker.GetCoins(std.CurrentRealm().Address()))
}

func Render(path string) string {
	var sb strings.Builder

	// Theme-specific header styling
	switch currentTheme {
	case "modern":
		// Modern theme - Clean and minimalist with emojis
		sb.WriteString(md.H1("🚀 Matija's Space"))
		sb.WriteString(md.Image(pfpCaption, pfp))
		sb.WriteString("\n")
		sb.WriteString(md.Italic(pfpCaption))
		sb.WriteString("\n")
		sb.WriteString(md.HorizontalRule())
		sb.WriteString(abtMe)
		sb.WriteString("\n")

	case "minimal":
		// Minimal theme - No emojis, minimal formatting
		sb.WriteString(md.H1("Matija Marjanovic"))
		sb.WriteString("\n")
		sb.WriteString(abtMe)
		sb.WriteString("\n")
		sb.WriteString(md.Image(pfpCaption, pfp))
		sb.WriteString("\n")
		sb.WriteString(pfpCaption)
		sb.WriteString("\n")

	default: // classic
		// Classic theme - Traditional blog style with decorative elements
		sb.WriteString(md.H1("✨ Welcome to Matija's Homepage ✨"))
		sb.WriteString("\n")
		sb.WriteString(md.Image(pfpCaption, pfp))
		sb.WriteString("\n")
		sb.WriteString(pfpCaption)
		sb.WriteString("\n")
		sb.WriteString(md.HorizontalRule())
		sb.WriteString(md.H2("About me"))
		sb.WriteString("\n")
		sb.WriteString(abtMe)
		sb.WriteString("\n")
	}

	// Theme-specific voting section
	switch currentTheme {
	case "modern":
		sb.WriteString(md.HorizontalRule())
		sb.WriteString(md.H2("🎨 Theme Selector"))
		sb.WriteString("Choose your preferred viewing experience:\n")
		items := []string{
			md.Link(ufmt.Sprintf("Modern Design (%d votes)", modernVotes), modernLink),
			md.Link(ufmt.Sprintf("Classic Style (%d votes)", classicVotes), classicLink),
			md.Link(ufmt.Sprintf("Minimal Look (%d votes)", minimalVotes), minimalLink),
		}
		sb.WriteString(md.BulletList(items))

	case "minimal":
		sb.WriteString("\n")
		sb.WriteString(md.H3("Theme Selection"))
		sb.WriteString(ufmt.Sprintf("Current theme: %s\n", currentTheme))
		sb.WriteString(ufmt.Sprintf("Votes - Modern: %d | Classic: %d | Minimal: %d\n",
			modernVotes, classicVotes, minimalVotes))
		sb.WriteString(md.Link("Modern", modernLink))
		sb.WriteString(" | ")
		sb.WriteString(md.Link("Classic", classicLink))
		sb.WriteString(" | ")
		sb.WriteString(md.Link("Minimal", minimalLink))
		sb.WriteString("\n")

	default: // classic
		sb.WriteString(md.HorizontalRule())
		sb.WriteString(md.H2("✨ Theme Customization ✨"))
		sb.WriteString(md.Bold("Choose Your Preferred Theme:"))
		sb.WriteString("\n\n")
		items := []string{
			ufmt.Sprintf("Modern 🚀 (%d votes) - %s", modernVotes, md.Link("Vote", modernLink)),
			ufmt.Sprintf("Classic ✨ (%d votes) - %s", classicVotes, md.Link("Vote", classicLink)),
			ufmt.Sprintf("Minimal ⚡ (%d votes) - %s", minimalVotes, md.Link("Vote", minimalLink)),
		}
		sb.WriteString(md.BulletList(items))
	}

	// Theme-specific footer/links section
	switch currentTheme {
	case "modern":
		sb.WriteString(md.HorizontalRule())
		sb.WriteString(md.Link("GitHub", "https://github.com/matijamarjanovic"))
		sb.WriteString(" | ")
		sb.WriteString(md.Link("LinkedIn", "https://www.linkedin.com/in/matijamarjanovic"))
		sb.WriteString("\n")

	case "minimal":
		sb.WriteString("\n")
		sb.WriteString(md.Link("GitHub", "https://github.com/matijamarjanovic"))
		sb.WriteString(" | ")
		sb.WriteString(md.Link("LinkedIn", "https://www.linkedin.com/in/matijamarjanovic"))
		sb.WriteString("\n")

	default: // classic
		sb.WriteString(md.HorizontalRule())
		sb.WriteString(md.H3("✨ Connect With Me"))
		items := []string{
			md.Link("🌟 GitHub", "https://github.com/matijamarjanovic"),
			md.Link("💼 LinkedIn", "https://www.linkedin.com/in/matijamarjanovic"),
		}
		sb.WriteString(md.BulletList(items))
	}

	return sb.String()
}

func UpdateModernLink(link string) {
	AssertAuthorized()
	modernLink = link
}

func UpdateClassicLink(link string) {
	AssertAuthorized()
	classicLink = link
}

func UpdateMinimalLink(link string) {
	AssertAuthorized()
	minimalLink = link
}
