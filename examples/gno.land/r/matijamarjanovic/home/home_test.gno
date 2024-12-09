package home

import (
        "std"
        "testing"

        "strings"

        "gno.land/p/demo/uassert"
        "gno.land/p/demo/urequire"
)

// Helper function to set up test environment
func setupTest() {
        std.TestSetOrigCaller(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"))
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

        std.TestSetOrigSend(coinsSent, coinsSpent)
        VoteModern()

        uassert.Equal(t, int64(75000000), modernVotes, "Modern votes should be calculated correctly")
        uassert.Equal(t, "modern", currentTheme, "Theme should be updated to modern")
}

func TestVoteClassic(t *testing.T) {
        setupTest()
        modernVotes, classicVotes, minimalVotes = 0, 0, 0

        coinsSent := std.NewCoins(std.NewCoin("ugnot", 75000000))
        coinsSpent := std.NewCoins(std.NewCoin("ugnot", 1))

        std.TestSetOrigSend(coinsSent, coinsSpent)
        VoteClassic()

        uassert.Equal(t, int64(75000000), classicVotes, "Classic votes should be calculated correctly")
        uassert.Equal(t, "classic", currentTheme, "Theme should be updated to classic")
}

func TestVoteMinimal(t *testing.T) {
        setupTest()
        modernVotes, classicVotes, minimalVotes = 0, 0, 0

        coinsSent := std.NewCoins(std.NewCoin("ugnot", 75000000))
        coinsSpent := std.NewCoins(std.NewCoin("ugnot", 1))

        std.TestSetOrigSend(coinsSent, coinsSpent)
        VoteMinimal()

        uassert.Equal(t, int64(75000000), minimalVotes, "Minimal votes should be calculated correctly")
        uassert.Equal(t, "minimal", currentTheme, "Theme should be updated to minimal")
}

func TestRender(t *testing.T) {
        setupTest()
        // Reset the state to known values
        modernVotes, classicVotes, minimalVotes = 0, 0, 0
        currentTheme = "classic"

        out := Render("")
        urequire.NotEqual(t, out, "", "Render output should not be empty")

        // Check if the output contains the expected theme-specific content
        expectedContent := "✨ Welcome to Matija's Homepage ✨"
        uassert.True(t, contains(out, expectedContent), "Render output should contain welcome message")
}

func contains(s, substr string) bool {
        return len(s) >= len(substr) && strings.Contains(s, substr)
}
