package home

import (
        "std"
        "testing"

        "gno.land/p/demo/uassert"
        "gno.land/p/demo/urequire"
)

func TestUpdatePFP(t *testing.T) {
        std.TestSetOrigCaller(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"))
        pfp = ""
        pfpCaption = ""

        UpdatePFP("https://example.com/pic.png", "New Caption")

        urequire.Equal(t, pfp, "https://example.com/pic.png", "Profile picture URL should be updated")
        urequire.Equal(t, pfpCaption, "New Caption", "Profile picture caption should be updated")
}

func TestUpdateAboutMe(t *testing.T) {
        std.TestSetOrigCaller(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"))
        abtMe = ""

        UpdateAboutMe("This is my new bio.")

        urequire.Equal(t, abtMe, "This is my new bio.", "About Me should be updated")
}

func TestVoteModern(t *testing.T) {
        modernVotes, classicVotes, minimalVotes = 0, 0, 0

        coinsSent := std.NewCoins(std.NewCoin("ugnot", 75000000))
        coinsSpent := std.NewCoins(std.NewCoin("ugnot", 1))

        std.TestSetOrigSend(coinsSent, coinsSpent)
        VoteModern()

        uassert.Equal(t, modernVotes, 75000000/1000000, "Modern votes should be calculated correctly")
        uassert.Equal(t, currentTheme, "modern", "Theme should be updated to modern")
}

func TestVoteClassic(t *testing.T) {
        modernVotes, classicVotes, minimalVotes = 0, 0, 0

        coinsSent := std.NewCoins(std.NewCoin("ugnot", 75000000))
        coinsSpent := std.NewCoins(std.NewCoin("ugnot", 1))

        std.TestSetOrigSend(coinsSent, coinsSpent)
        VoteClassic()

        uassert.Equal(t, classicVotes, 75000000/1000000, "Classic votes should be calculated correctly")
        uassert.Equal(t, currentTheme, "classic", "Theme should be updated to classic")
}

func TestVoteMinimal(t *testing.T) {
        modernVotes, classicVotes, minimalVotes = 0, 0, 0

        coinsSent := std.NewCoins(std.NewCoin("ugnot", 75000000))
        coinsSpent := std.NewCoins(std.NewCoin("ugnot", 1))

        std.TestSetOrigSend(coinsSent, coinsSpent)
        VoteMinimal()

        uassert.Equal(t, minimalVotes, 75000000/1000000, "Minimal votes should be calculated correctly")
        uassert.Equal(t, currentTheme, "minimal", "Theme should be updated to minimal")
}

func TestRender(t *testing.T) {
        redVotes, blueVotes, greenVotes = 0, 0, 0
        currentColor = `rgb(120, 120, 120)`

        out := Render("")

        urequire.NotEqual(t, out, "", "Render output should not be empty")

        expectedString := "<div style='font-family: \"Trebuchet MS\", sans-serif;font-size: 5rem;"
        uassert.True(t, contains(out, expectedString), "Render output should contain expected string")
}

func TestUpdateBlueLink(t *testing.T) {
        std.TestSetOrigCaller(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"))
        blueLink = ""

        UpdateBlueLink("https://example.com/blue")

        if blueLink != "https://example.com/blue" {
                t.Fatalf("expected blueLink to be https://example.com/blue, got %s", blueLink)
        }
}

func TestUpdateRedLink(t *testing.T) {
        std.TestSetOrigCaller(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"))
        redLink = ""

        UpdateRedLink("https://example.com/red")

        if redLink != "https://example.com/red" {
                t.Fatalf("expected redLink to be https://example.com/red, got %s", redLink)
        }
}

func TestUpdateGreenLink(t *testing.T) {
        std.TestSetOrigCaller(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"))
        greenLink = ""

        UpdateGreenLink("https://example.com/green")

        if greenLink != "https://example.com/green" {
                t.Fatalf("expected greenLink to be https://example.com/green, got %s", greenLink)
        }
}

func contains(s, substr string) bool {
        return len(s) >= len(substr) && s[0:len(substr)] == substr
}
