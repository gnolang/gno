package rps

import (
	"errors"
	"std"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
	"gno.land/p/lou/ascii"
	"gno.land/p/lou/tokenaccess"
)

const (
	Rock     = "rock"
	Paper    = "paper"
	Scissors = "scissors"
)

var (
	hands = []string{
		Rock,
		Paper,
		Scissors,
	}
	winMap = map[string]string{
		Rock:     Scissors,
		Paper:    Rock,
		Scissors: Paper,
	}

	ErrInvalidHand = errors.New("invalid hand input: only rock, paper or scissors")
)

type Profile struct {
	wins    int64
	losses  int64
	account string
	token   *grc20.Token
	led     *grc20.PrivateLedger
	owner   *std.Address
}

type GameSession struct {
	hasWon bool
	isATie bool
	detail string
}

func (p Profile) Wins() int64 {
	return p.wins
}

func (p Profile) Losses() int64 {
	return p.losses
}

func (p Profile) Account() string {
	return p.account
}

func (p *Profile) SetWins(wins int64) {
	p.wins = wins
}

func (p *Profile) SetLosses(losses int64) {
	p.losses = losses
}

func (p *Profile) SetAddrOwner(addr std.Address) {
	p.owner = &addr
}

func (p Profile) RewardOnWin(owner std.Address) {
	tokenaccess.RewardAndApprove(owner, 10, p.token, p.led, std.Address(p.account), 5)
}

func (p Profile) String() string {
	out := ""

	if p.Wins() > 0 || p.Losses() > 0 {
		out += "Career: "
	}
	out += ascii.Repeat("🏆", int(p.Wins()))
	out += ascii.Repeat("❌", int(p.Losses()))
	return out
}

func NewProfile(
	account string,
	token *grc20.Token,
	led *grc20.PrivateLedger,
) *Profile {
	return &Profile{
		wins:    0,
		losses:  0,
		account: account,
		token:   token,
		led:     led,
		owner:   nil,
	}
}

func NewProfileAndOwner(
	account string,
	token *grc20.Token,
	led *grc20.PrivateLedger,
	owner std.Address,
) *Profile {
	return &Profile{
		wins:    0,
		losses:  0,
		account: account,
		token:   token,
		led:     led,
		owner:   &owner,
	}
}

func NewSession() *GameSession {
	return &GameSession{
		hasWon: false,
		isATie: false,
		detail: "",
	}
}

func (g GameSession) IsATie() bool {
	return g.isATie
}

func (g GameSession) HasWon() bool {
	return g.hasWon
}

func (g *GameSession) SendHand(input string) error {
	if _, ok := winMap[input]; !ok {
		return ErrInvalidHand
	}

	opponent := GetHand()
	g.isATie = false
	g.hasWon = false

	if input == opponent {
		g.detail = "It's a **tie**! Try Again!\n"
		g.isATie = true
	} else if winMap[input] == opponent {
		g.detail = ufmt.Sprintf("**%s** beats **%s**! You Won!\n", input, opponent)
		g.hasWon = true
	} else {
		g.detail = ufmt.Sprintf("**%s** beats **%s**! You Lost!\n", opponent, input)
	}
	return nil
}

func (g GameSession) String() string {
	return g.detail
}

func GetHand() string {
	random := int(std.ChainHeight()) % len(hands)
	return hands[random]
}
