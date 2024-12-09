package home

import (
        "std"

        "gno.land/p/demo/ufmt"
        "gno.land/r/leon/hof"
        "gno.land/r/matijamarjanovic/mdui"
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
        abtMe =
                `### About me
Motivated Computer Science student with strong
analytical and problem-solving skills. Proficient in
programming and version control, with a high level of
focus and attention to detail. Eager to apply academic
knowledge to real-world projects and contribute to
innovative technology solutions.

In addition to my academic pursuits, I enjoy traveling and staying active through weightlifting. I have a keen interest in electronic music and often explore various genres. I believe in maintaining a balanced lifestyle that complements my professional development.`        
        modernVotes = 0
        classicVotes = 0
        minimalVotes = 0
        currentTheme = "classic"
        modernLink = "https://www.google.com"
        classicLink = "https://www.google.com"
        minimalLink = "https://www.google.com"
        hof.Register()
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
        ugnotAmount := std.GetOrigSend().AmountOf("ugnot")
        votes := ugnotAmount
        modernVotes += votes
        updateCurrentTheme()
}

func VoteClassic() {
        ugnotAmount := std.GetOrigSend().AmountOf("ugnot")
        votes := ugnotAmount
        classicVotes += votes
        updateCurrentTheme()
}

func VoteMinimal() {
        ugnotAmount := std.GetOrigSend().AmountOf("ugnot")
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

        banker := std.GetBanker(std.BankerTypeRealmSend)
        ownerAddr := Address()

        banker.SendCoins(std.CurrentRealm().Addr(), ownerAddr, banker.GetCoins(std.CurrentRealm().Addr()))
}

func Render(path string) string {
        builder := mdui.NewBuilder()

        // Theme-specific header styling
        switch currentTheme {
        case "modern":
                // Modern theme - Clean and minimalist with emojis
                builder.AddHeading(1, "🚀 Matija's Space")
                builder.AddBreakLine()
                builder.AddImage(pfp, "Profile Picture")
                builder.AddBreakLine()
                builder.AddParagraph("_" + pfpCaption + "_") // Italicized caption
                builder.AddDivider()
                builder.AddParagraph(abtMe)

        case "minimal":
                // Minimal theme - No emojis, minimal formatting
                builder.AddHeading(1, "Matija Marjanovic")
                builder.AddBreakLine()
                builder.AddParagraph(abtMe)
                builder.AddBreakLine()
                builder.AddImage(pfp, "Profile Picture")
                builder.AddBreakLine()
                builder.AddParagraph(pfpCaption)

        default: // classic
                // Classic theme - Traditional blog style with decorative elements
                builder.AddHeading(1, "✨ Welcome to Matija's Homepage ✨")
                builder.AddBreakLine()
                builder.AddImage(pfp, "Profile Picture")
                builder.AddBreakLine()
                builder.AddParagraph("**" + pfpCaption + "**") // Bold caption
                builder.AddDivider()
                builder.AddParagraph(abtMe)
        }

        // Theme-specific voting section
        switch currentTheme {
        case "modern":
                builder.AddDivider()
                builder.AddHeading(2, "🎨 Theme Selector")
                builder.AddParagraph("Choose your preferred viewing experience:")
                builder.AddList([]string{
                        ufmt.Sprintf("Modern Design (%d votes)", modernVotes),
                        ufmt.Sprintf("Classic Style (%d votes)", classicVotes),
                        ufmt.Sprintf("Minimal Look (%d votes)", minimalVotes),
                }, []string{modernLink, classicLink, minimalLink}, false)

        case "minimal":
                builder.AddBreakLine()
                builder.AddHeading(3, "Theme Selection")
                builder.AddParagraph(ufmt.Sprintf("Current theme: %s", currentTheme))
                builder.AddParagraph(ufmt.Sprintf("Votes - Modern: %d | Classic: %d | Minimal: %d",
                        modernVotes, classicVotes, minimalVotes))
                builder.AddButton("Modern", modernLink)
                builder.AddButton("Classic", classicLink)
                builder.AddButton("Minimal", minimalLink)

        default: // classic
                builder.AddDivider()
                builder.AddHeading(2, "✨ Theme Customization ✨")
                builder.AddParagraph("**Choose Your Preferred Theme:**")
                builder.AddTable(
                        []string{"Theme", "Votes", "Action"},
                        [][]string{
                                {"Modern 🚀", ufmt.Sprintf("%d", modernVotes), "[Vote](" + modernLink + ")"},
                                {"Classic ✨", ufmt.Sprintf("%d", classicVotes), "[Vote](" + classicLink + ")"},
                                {"Minimal ⚡", ufmt.Sprintf("%d", minimalVotes), "[Vote](" + minimalLink + ")"},
                        },
                )
        }

        // Theme-specific footer/links section
        switch currentTheme {
        case "modern":
                builder.AddDivider()
                builder.AddBadgeWithIcon("GitHub", "black", "github")
                builder.AddBadgeWithIcon("LinkedIn", "blue", "linkedin")

        case "minimal":
                builder.AddBreakLine()
                builder.AddNavbar([]mdui.NavLink{
                        {Text: "GitHub", Href: "https://github.com/matijamarjanovic"},
                        {Text: "LinkedIn", Href: "https://www.linkedin.com/in/matijamarjanovic"},
                })

        default: // classic
                builder.AddDivider()
                builder.AddHeading(3, "✨ Connect With Me")
                builder.AddNavbar([]mdui.NavLink{
                        {Text: "🌟 GitHub", Href: "https://github.com/matijamarjanovic"},
                        {Text: "💼 LinkedIn", Href: "https://www.linkedin.com/in/matijamarjanovic"},
                })
        }

        return builder.Render()
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
