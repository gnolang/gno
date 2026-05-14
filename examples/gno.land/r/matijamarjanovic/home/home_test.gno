package home

import (
	"chain"
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestUpdatePFP(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")))
	pfp = ""
	pfpCaption = ""

	UpdatePFP(cross, "https://example.com/pic.png", "New Caption")

	urequire.Equal(t, pfp, "https://example.com/pic.png", "Profile picture URL should be updated")
	urequire.Equal(t, pfpCaption, "New Caption", "Profile picture caption should be updated")
}

func TestUpdateAboutMe(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")))
	abtMe = ""

	UpdateAboutMe(cross, "This is my new bio.")

	urequire.Equal(t, abtMe, "This is my new bio.", "About Me should be updated")
}

func TestVoteModern(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")))
	modernVotes, classicVotes, minimalVotes = 0, 0, 0

	coinsSent := chain.NewCoins(chain.NewCoin("ugnot", 75000000))
	coinsSpent := chain.NewCoins(chain.NewCoin("ugnot", 1))

	testing.SetOriginSend(coinsSent)
	testing.SetOriginSpend(coinsSpent)
	VoteModern(cross)

	uassert.Equal(t, int64(75000000), modernVotes, "Modern votes should be calculated correctly")
	uassert.Equal(t, "modern", currentTheme, "Theme should be updated to modern")
}

func TestVoteClassic(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")))
	modernVotes, classicVotes, minimalVotes = 0, 0, 0

	coinsSent := chain.NewCoins(chain.NewCoin("ugnot", 75000000))
	coinsSpent := chain.NewCoins(chain.NewCoin("ugnot", 1))

	testing.SetOriginSend(coinsSent)
	testing.SetOriginSpend(coinsSpent)
	VoteClassic(cross)

	uassert.Equal(t, int64(75000000), classicVotes, "Classic votes should be calculated correctly")
	uassert.Equal(t, "classic", currentTheme, "Theme should be updated to classic")
}

func TestVoteMinimal(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")))
	modernVotes, classicVotes, minimalVotes = 0, 0, 0

	coinsSent := chain.NewCoins(chain.NewCoin("ugnot", 75000000))
	coinsSpent := chain.NewCoins(chain.NewCoin("ugnot", 1))

	testing.SetOriginSend(coinsSent)
	testing.SetOriginSpend(coinsSpent)
	VoteMinimal(cross)

	uassert.Equal(t, int64(75000000), minimalVotes, "Minimal votes should be calculated correctly")
	uassert.Equal(t, "minimal", currentTheme, "Theme should be updated to minimal")
}

func TestRender(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")))
	modernVotes, classicVotes, minimalVotes = 0, 0, 0
	currentTheme = "classic"
	pfp = "https://example.com/pic.png"
	pfpCaption = "Test Caption"
	abtMe = "Test About Me"

	out := Render("")
	urequire.NotEqual(t, out, "", "Render output should not be empty")

	uassert.True(t, strings.Contains(out, "✨ Welcome to Matija's Homepage ✨"), "Classic theme should have correct header")
	uassert.True(t, strings.Contains(out, pfp), "Should contain profile picture URL")
	uassert.True(t, strings.Contains(out, pfpCaption), "Should contain profile picture caption")
	uassert.True(t, strings.Contains(out, "About me"), "Should contain About me section")
	uassert.True(t, strings.Contains(out, abtMe), "Should contain about me content")
	uassert.True(t, strings.Contains(out, "Theme Customization"), "Should contain theme customization section")
	uassert.True(t, strings.Contains(out, "Connect With Me"), "Should contain connect section")
}

func TestRenderModernTheme(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")))
	modernVotes, classicVotes, minimalVotes = 100, 0, 0
	currentTheme = "modern"
	updateCurrentTheme()

	out := Render("")
	uassert.True(t, strings.Contains(out, "🚀 Matija's Space"), "Modern theme should have correct header")
}

func TestRenderMinimalTheme(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")))
	modernVotes, classicVotes, minimalVotes = 0, 0, 100
	currentTheme = "minimal"
	updateCurrentTheme()

	out := Render("")
	uassert.True(t, strings.Contains(out, "Matija Marjanovic"), "Minimal theme should have correct header")
}
