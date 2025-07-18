package home

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/lou/eightball"
	"gno.land/p/lou/rps"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"

	"gno.land/r/sys/users"
)

var (
	profiles        *avl.Tree
	currentProfile  *rps.Profile
	gameSession     *rps.GameSession
	eightballAnswer string
)

func init() {
	currentProfile = rps.NewProfile("", token, ledger)
	gameSession = rps.NewSession()
	profiles = avl.NewTree()
}

// ROCK PAPER SCISSORS

func Play(_ realm, input string) {
	caller := std.PreviousRealm().Address()
	addr := caller.String()

	value, ok := profiles.Get(addr)
	var profile *rps.Profile

	if ok {
		profile, _ = value.(*rps.Profile)
	}
	if profile == nil {
		profile = rps.NewProfile(addr, token, ledger)
		profiles.Set(addr, profile)
	}
	currentProfile = profile

	if err := gameSession.SendHand(input); err != nil {
		panic(err)
	}

	if gameSession.IsATie() {
		return
	}

	if gameSession.HasWon() {
		profile.SetWins(profile.Wins() + 1)
		profile.RewardOnWin(admin)
	} else {
		profile.SetLosses(profile.Losses() + 1)
	}
}

func RenderRockPaperScissors() string {
	user := users.ResolveAddress(std.Address(currentProfile.Account()))
	out := md.H2("Rock, Paper, Scissor\n\n")

	out += "(This game will reward you some LTK (LouTokens)!)\n\n"
	out += currentProfile.String() + "\n\n"
	if user != nil && currentProfile.Account() != "" {
		out += md.Italic("--> " + user.RenderLink("") + "\n\n")
	} else if currentProfile.Account() != "" {
		out += md.Italic("--> "+currentProfile.Account()) + "\n\n"
	}

	out += "Play [ "
	out += md.Link("Rock", txlink.Call("Play", "input", "rock")) + " | "
	out += md.Link("Paper", txlink.Call("Play", "input", "paper")) + " | "
	out += md.Link("Scissors ", txlink.Call("Play", "input", "scissors")) + " ]\n\n"
	out += gameSession.String() + "\n\n"

	out += ufmt.Sprintf("current allowance: **%d (%s)**\n\n", token.Allowance(admin, std.Address(currentProfile.Account())), token.GetSymbol())
	return out
}

// MAGIC 8 BALL

func Ask(_ realm, question string) string {
	answer := eightball.GetResponse(question)
	eightballAnswer = answer + "\n\n"
	return answer
}

func renderMagic8Ball() string {
	out := md.H2("Magic 8 Ball\n\n")
	out += md.Link("[ ASK A QUESTION ]", txlink.Call("Ask", "question", "Question")) + "\n\n"
	out += eightballAnswer
	return out
}
