package home

import ("testing"
		"std"
)


func TestUpdatePFP(t *testing.T) {
	std.TestSetOrigCaller(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"))
    pfp = ""
    pfpCaption = ""

    UpdatePFP("https://example.com/pic.png", "New Caption")

    if pfp != "https://example.com/pic.png" {
        t.Fatalf("expected pfp to be https://example.com/pic.png, got %s", pfp)
    }

    if pfpCaption != "New Caption" {
        t.Fatalf("expected caption to be 'New Caption', got %s", pfpCaption)
    }
}

func TestUpdateAboutMe(t *testing.T) {
	std.TestSetOrigCaller(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y"))

    abtMe = ""

    UpdateAboutMe("This is my new bio.")

    if abtMe != "This is my new bio." {
        t.Fatalf("expected About Me to be updated, got %s", abtMe)
    }
}

func TestVoteRed(t *testing.T) {
    redVotes, blueVotes, greenVotes = 0, 0, 0

	coinsSent := std.NewCoins(std.NewCoin("ugnot", 75000000))
	coinsSpent := std.NewCoins(std.NewCoin("ugnot", 1))

    std.TestSetOrigSend(coinsSent, coinsSpent) 
    VoteRed()

    if redVotes != 75000000/750 {
        t.Fatalf("expected redVotes to be %d, got %d", 75000000/750, redVotes)
    }

    red, _, _ := calculateRGB(redVotes, greenVotes, blueVotes)
    if red != 255 {
        t.Fatalf("expected red value to be 255, got %d", red)
    }
}

func TestVoteBlue(t *testing.T) {
    redVotes, blueVotes, greenVotes = 0, 0, 0

	coinsSent := std.NewCoins(std.NewCoin("ugnot", 75000000))
	coinsSpent := std.NewCoins(std.NewCoin("ugnot", 1))

    std.TestSetOrigSend(coinsSent, coinsSpent)
    VoteBlue()

    if blueVotes != 75000000/750 {
        t.Fatalf("expected blueVotes to be %d, got %d", 75000000/750, blueVotes)
    }

    _, _, blue := calculateRGB(redVotes, greenVotes, blueVotes)
    if blue != 255 {
        t.Fatalf("expected blue value to be 255, got %d", blue)
    }
}

func TestVoteGreen(t *testing.T) {
    redVotes, greenVotes, greenVotes = 0, 0, 0

	coinsSent := std.NewCoins(std.NewCoin("ugnot", 75000000))
	coinsSpent := std.NewCoins(std.NewCoin("ugnot", 1))

    std.TestSetOrigSend(coinsSent, coinsSpent) 
    VoteGreen()

    if greenVotes != 75000000/750 {
        t.Fatalf("expected greenVotes to be %d, got %d", 75000000/750, greenVotes)
    }

    // Verify current color calculation
    _, green, _ := calculateRGB(redVotes, greenVotes, greenVotes)
    if green != 255 {
        t.Fatalf("expected green value to be 255, got %d", green)
    }
}

func TestRender(t *testing.T) {
    redVotes, blueVotes, greenVotes = 0, 0, 0
    currentColor = `rgb(120, 120, 120)`

    out := Render("")

    if out == "" {
        t.Fatalf("Render output is empty")
    }

    expectedString := "<div style='font-family: \"Trebuchet MS\", sans-serif;font-size: 5rem;"
    if !contains(out, expectedString) {
        t.Fatalf("expected output to contain %s, got %s", expectedString, out)
    }
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
