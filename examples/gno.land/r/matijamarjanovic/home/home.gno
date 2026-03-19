package home

import (
	"chain/banker"
	"chain/runtime"
	"strings"

	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/leon/hor"
)

var (
	pfp        string
	pfpCaption string
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
	pfp = "https://avatars.githubusercontent.com/u/93043005?s=60&v=4"
	pfpCaption = "My github profile picture"
	abtMe = `Blockchain & Full Stack Developer and a Computer Science Student who enjoys hackathons and building cool stuff. My current passion is DeFi. Outside of work, I enjoy travelling, weightlifting and combat sports.`

	modernVotes = 0
	classicVotes = 0
	minimalVotes = 0
	currentTheme = "classic"
	modernLink = txlink.NewLink("VoteModern").URL()
	classicLink = txlink.NewLink("VoteClassic").URL()
	minimalLink = txlink.NewLink("VoteMinimal").URL()
	hor.Register(cross, "Matija Marijanovic's Home Realm", "")
}

func UpdatePFP(cur realm, url, caption string) {
	if err := Auth.DoByPrevious("update_pfp", func() error {
		pfp = url
		pfpCaption = caption
		return nil
	}); err != nil {
		panic(err)
	}
}

func UpdateAboutMe(cur realm, col1 string) {
	if err := Auth.DoByPrevious("update_about_me", func() error {
		abtMe = col1
		return nil
	}); err != nil {
		panic(err)
	}
}

func VoteModern(cur realm) {
	ugnotAmount := banker.OriginSend().AmountOf("ugnot")
	votes := ugnotAmount
	modernVotes += votes
	updateCurrentTheme()
}

func VoteClassic(cur realm) {
	ugnotAmount := banker.OriginSend().AmountOf("ugnot")
	votes := ugnotAmount
	classicVotes += votes
	updateCurrentTheme()
}

func VoteMinimal(cur realm) {
	ugnotAmount := banker.OriginSend().AmountOf("ugnot")
	votes := ugnotAmount
	minimalVotes += votes
	updateCurrentTheme()
}

func CollectBalance(cur realm) {
	if err := Auth.DoByPrevious("collect_balance", func() error {
		banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
		ownerAddr := Address()
		banker_.SendCoins(runtime.CurrentRealm().Address(), ownerAddr, banker_.GetCoins(runtime.CurrentRealm().Address()))
		return nil
	}); err != nil {
		panic(err)
	}
}

func Render(path string) string {
	var sb strings.Builder

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

	default:
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

	default:
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

	default:
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
