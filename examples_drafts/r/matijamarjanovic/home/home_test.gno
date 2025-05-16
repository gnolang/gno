package home

import (
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

// Helper function to set up test environment
func setupTest() {
	testing.SetOriginCaller(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"))
}

func TestUpdatePFP(t *testing.T) {
	setupTest()
	pfp = ""
	pfpCaption = ""

	UpdatePFP("https://example.com/pic.png", "New Caption")

	urequire.Equal(t, pfp, "https://example.com/pic.png", "Profile picture URL should be updated")
	urequire.Equal(t, pfpCaption, "New Caption", "Profile picture caption should be updated")
}

func TestUpdateAboutMe(t *testing.T) {
	setupTest()
	abtMe = ""

	UpdateAboutMe("This is my new bio.")

	urequire.Equal(t, abtMe, "This is my new bio.", "About Me should be updated")
}

func TestVoteModern(t *testing.T) {
	setupTest()
	modernVotes, classicVotes, minimalVotes = 0, 0, 0

	coinsSent := std.NewCoins(std.NewCoin("ugnot", 75000000))
	coinsSpent := std.NewCoins(std.NewCoin("ugnot", 1))

	testing.SetOriginSend(coinsSent)
	testing.SetOriginSpend(coinsSpent)
	VoteModern()

	uassert.Equal(t, int64(75000000), modernVotes, "Modern votes should be calculated correctly")
	uassert.Equal(t, "modern", currentTheme, "Theme should be updated to modern")
}

func TestVoteClassic(t *testing.T) {
	setupTest()
	modernVotes, classicVotes, minimalVotes = 0, 0, 0

	coinsSent := std.NewCoins(std.NewCoin("ugnot", 75000000))
	coinsSpent := std.NewCoins(std.NewCoin("ugnot", 1))

	testing.SetOriginSend(coinsSent)
	testing.SetOriginSpend(coinsSpent)
	VoteClassic()

	uassert.Equal(t, int64(75000000), classicVotes, "Classic votes should be calculated correctly")
	uassert.Equal(t, "classic", currentTheme, "Theme should be updated to classic")
}

func TestVoteMinimal(t *testing.T) {
	setupTest()
	modernVotes, classicVotes, minimalVotes = 0, 0, 0

	coinsSent := std.NewCoins(std.NewCoin("ugnot", 75000000))
	coinsSpent := std.NewCoins(std.NewCoin("ugnot", 1))

	testing.SetOriginSend(coinsSent)
	testing.SetOriginSpend(coinsSpent)
	VoteMinimal()

	uassert.Equal(t, int64(75000000), minimalVotes, "Minimal votes should be calculated correctly")
	uassert.Equal(t, "minimal", currentTheme, "Theme should be updated to minimal")
}

func TestRender(t *testing.T) {
	setupTest()
	// Reset the state to known values
	modernVotes, classicVotes, minimalVotes = 0, 0, 0
	currentTheme = "classic"
	pfp = "https://example.com/pic.png"
	pfpCaption = "Test Caption"
	abtMe = "Test About Me"

	out := Render("")
	urequire.NotEqual(t, out, "", "Render output should not be empty")

	// Test classic theme specific content
	uassert.True(t, strings.Contains(out, "✨ Welcome to Matija's Homepage ✨"), "Classic theme should have correct header")
	uassert.True(t, strings.Contains(out, pfp), "Should contain profile picture URL")
	uassert.True(t, strings.Contains(out, pfpCaption), "Should contain profile picture caption")
	uassert.True(t, strings.Contains(out, "About me"), "Should contain About me section")
	uassert.True(t, strings.Contains(out, abtMe), "Should contain about me content")
	uassert.True(t, strings.Contains(out, "Theme Customization"), "Should contain theme customization section")
	uassert.True(t, strings.Contains(out, "Connect With Me"), "Should contain connect section")
}

func TestRenderModernTheme(t *testing.T) {
	setupTest()
	modernVotes, classicVotes, minimalVotes = 100, 0, 0
	currentTheme = "modern"
	updateCurrentTheme()

	out := Render("")
	uassert.True(t, strings.Contains(out, "🚀 Matija's Space"), "Modern theme should have correct header")
}

func TestRenderMinimalTheme(t *testing.T) {
	setupTest()
	modernVotes, classicVotes, minimalVotes = 0, 0, 100
	currentTheme = "minimal"
	updateCurrentTheme()

	out := Render("")
	uassert.True(t, strings.Contains(out, "Matija Marjanovic"), "Minimal theme should have correct header")
}

func TestUpdateLinks(t *testing.T) {
	setupTest()

	newLink := "https://example.com/vote"

	UpdateModernLink(newLink)
	urequire.Equal(t, modernLink, newLink, "Modern link should be updated")

	UpdateClassicLink(newLink)
	urequire.Equal(t, classicLink, newLink, "Classic link should be updated")

	UpdateMinimalLink(newLink)
	urequire.Equal(t, minimalLink, newLink, "Minimal link should be updated")
}
