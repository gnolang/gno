package diceduel

import (
	"errors"
	"math/rand"
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/entropy"
	"gno.land/p/demo/seqid"
	"gno.land/r/demo/users"
)

type game struct {
	player1, player2 std.Address
	roll1, roll2     int
	score1, score2   int
	round, maxRounds int
}

var (
	games avl.Tree
	id    seqid.ID
	seed  = uint64(entropy.New().Seed())
)

// Roll a dice between 1 and 6
func rollDice() int {
	r := rand.New(rand.NewPCG(seed, 0xdeadbeef))

	return r.IntN(6) + 1
}

// NewGame initializes a new Dice Duel game with the specified number of rounds
func NewGame(player std.Address, rounds int) int {
	if rounds <= 0 {
		panic("invalid number of rounds")
	}

	gameID := id.Next()
	games.Set(gameID.String(), &game{
		player1:   std.PrevRealm().Addr(),
		player2:   player,
		maxRounds: rounds,
	})
	return int(gameID)
}

// Play simulates a dice roll for the current player and updates the game state
func Play(idx int) int {
	g, err := gameByID(idx)
	if err != nil {
		panic(err)
	}

	roll := rollDice()

	if err := g.play(std.PrevRealm().Addr(), roll); err != nil {
		panic(err)
	}

	// Check if both players have played, then advance the round
	if g.roll1 != 0 && g.roll2 != 0 {
		g.round++
		g.roll1, g.roll2 = 0, 0 // Reset rolls for the next round
	}

	return roll
}

// play handles a player's dice roll and updates their score
func (g *game) play(player std.Address, roll int) error {
	if player != g.player1 && player != g.player2 {
		return errors.New("invalid player")
	}

	if g.round >= g.maxRounds {
		return errors.New("game over")
	}

	if player == g.player1 && g.roll1 == 0 {
		g.roll1 = roll
		g.score1 += roll
		return nil
	}

	if player == g.player2 && g.roll2 == 0 {
		g.roll2 = roll
		g.score2 += roll
		return nil
	}

	return errors.New("already played")
}

// winner determines the winner of the game based on scores
func (g *game) winner() string {
	if g.round < g.maxRounds {
		return "Game still in progress"
	}

	if g.score1 > g.score2 {
		return "Player 1 wins!"
	} else if g.score2 > g.score1 {
		return "Player 2 wins!"
	} else {
		return "It's a draw!"
	}
}

// diceIcon returns a string representation of the dice roll
func diceIcon(roll int) string {
	switch roll {
	case 1:
		return "🎲1"
	case 2:
		return "🎲2"
	case 3:
		return "🎲3"
	case 4:
		return "🎲4"
	case 5:
		return "🎲5"
	case 6:
		return "🎲6"
	default:
		return strconv.Itoa(roll)
	}
}

// Render provides a summary of the current state of all games
func Render(path string) string {
	output := `# 🎲 Dice Duel

## Game Rules:
Dice Duel is a two-player game where each player rolls a dice over a specified number of rounds
1. **Create a game** with an opponent using [NewGame](dice?help&__func=NewGame)
2. **Play rounds** using [Play](dice?help&__func=Play), where each player rolls a dice
   - Dice values range from 1 to 6 for each roll
   - Each player rolls the dice in turn, and the rolls are recorded for each round
## Actions:
* [NewGame](dice?help&__func=NewGame) opponentAddress rounds
  - **opponentAddress**: The address of the opponent player
  - **rounds**: The number of rounds for the game. Must be a positive integer
* [Play](dice?help&__func=Play) gameIndex
  - **gameIndex**: The unique index of the game you want to play
## Recent Games:
Displays a summary of the most recent games. Up to 10 recent games are shown

| Game | Player 1 | Roll 1 | Score 1 | Player 2 | Roll 2 | Score 2 | Round | Max Rounds | Winner |
|------|----------|--------|---------|----------|--------|---------|-------|------------|--------|
`

	// Display up to 10 recent games
	maxGames := 10
	for n := int(id); n > 0 && int(id)-n < maxGames; n-- {
		g, err := gameByID(n)
		if err != nil {
			continue
		}

		output += strconv.Itoa(n) + " | " +
			shortName(g.player1) + " | " + diceIcon(g.roll1) + " | " + strconv.Itoa(g.score1) + " | " +
			shortName(g.player2) + " | " + diceIcon(g.roll2) + " | " + strconv.Itoa(g.score2) + " | " +
			strconv.Itoa(g.round) + " | " + strconv.Itoa(g.maxRounds) + " | " +
			g.winner() + "\n"
	}
	return output
}

// shortName returns a shortened name for the given address
func shortName(addr std.Address) string {
	user := users.GetUserByAddress(addr)
	if user != nil {
		return user.Name
	}
	if len(addr) < 10 {
		return string(addr)
	}
	return string(addr)[:10] + "..."
}

// gameByID retrieves the game state by its ID
func gameByID(idx int) (*game, error) {
	v, ok := games.Get(seqid.ID(idx).String())
	if !ok {
		return nil, errors.New("game not found")
	}
	return v.(*game), nil
}
