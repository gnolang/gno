package diceroller

import (
	"errors"
	"math/rand"
	"sort"
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/entropy"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"
)

type (
	// game represents a Dice Roller game between two players
	game struct {
		player1, player2 std.Address
		roll1, roll2     int
	}

	// player holds the information about each player including their stats
	player struct {
		addr                              std.Address
		wins, losses, draws, points, rank int
	}

	// leaderBoard is a slice of players, used to sort players by rank
	leaderBoard []player
)

const (
	// Constants to represent game result outcomes
	empty = iota
	win
	draw
	lose
)

var (
	games  avl.Tree // AVL tree for storing game states
	gameId seqid.ID // Sequence ID for games

	players avl.Tree // AVL tree for storing player data

	seed = uint64(entropy.New().Seed())
)

// rollDice generates a random dice roll between 1 and 6
func rollDice() int {
	r := rand.New(rand.NewPCG(seed, 0xdeadbeef))
	return r.IntN(6) + 1
}

// NewGame initializes a new game with the provided opponent's address
func NewGame(addr std.Address) int {
	if !addr.IsValid() {
		panic("invalid opponent's address")
	}

	games.Set(gameId.Next().String(), &game{
		player1: std.PrevRealm().Addr(),
		player2: addr,
	})

	return int(gameId)
}

// Play allows a player to roll the dice and updates the game state accordingly
func Play(idx int) int {
	g, err := getGame(idx)
	if err != nil {
		panic(err)
	}

	roll := rollDice() // Random the player's dice roll

	// Play the game and update the player's roll
	if err := g.play(std.PrevRealm().Addr(), roll); err != nil {
		panic(err)
	}

	// If both players have rolled, update the results and leaderboard
	if g.roll1 != 0 && g.roll2 != 0 {
		// If the player is playing against themselves, no points are awarded
		if g.player1 == g.player2 {
			return roll
		}

		player1 := getPlayer(g.player1)
		player2 := getPlayer(g.player2)

		if g.roll1 > g.roll2 {
			player1.updateResult(win)
			player2.updateResult(lose)
		} else if g.roll2 > g.roll1 {
			player2.updateResult(win)
			player1.updateResult(lose)
		} else {
			player1.updateResult(draw)
			player2.updateResult(draw)
		}
	}

	return roll
}

// play processes a player's roll and updates their score
func (g *game) play(player std.Address, roll int) error {
	if player != g.player1 && player != g.player2 {
		return errors.New("invalid player")
	}

	if g.roll1 != 0 && g.roll2 != 0 {
		return errors.New("game over")
	}

	if player == g.player1 && g.roll1 == 0 {
		g.roll1 = roll
		return nil
	}

	if player == g.player2 && g.roll2 == 0 {
		g.roll2 = roll
		return nil
	}

	return errors.New("already played")
}

// winner determines the winner based on scores
func (g *game) winner() string {
	if g.roll1 == 0 || g.roll2 == 0 {
		return "Game still in progress"
	}

	if g.roll1 > g.roll2 {
		return "Player 1 wins!"
	} else if g.roll2 > g.roll1 {
		return "Player 2 wins!"
	} else {
		return "It's a draw!"
	}
}

// diceIcon returns an icon of the dice roll
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

// Render provides a summary of the current state of games and leader board
func Render(path string) string {
	output := `# 🎲 Dice Roller

## Game Rules:
Dice Roller is a two-player game where each player rolls a dice once
1. **Create a game** with an opponent using [NewGame](diceroller?help&__func=NewGame)
2. **Play** using [Play](diceroller?help&__func=Play), where each player rolls a dice

## Scoring:
- Playing against yourself? No points or stats changes for you
- **Win:** 3 points
- **Loss:** 0 points
- **Draw:** 1 point each

## Actions:
* [NewGame](diceroller?help&__func=NewGame) opponentAddress
  - **opponentAddress**: The address of the opponent player
* [Play](diceroller?help&__func=Play) gameIndex
  - **gameIndex**: The unique index of the game you want to play

## Recent Games:
Displays a summary of the most recent games. Up to 10 recent games are shown

| Game | Player 1 | Roll 1 | Player 2 | Roll 2 | Winner |
|------|----------|--------|----------|--------|--------|
`

	// Display up to 10 recent games
	maxGames := 10
	for n := int(gameId); n > 0 && int(gameId)-n < maxGames; n-- {
		g, err := getGame(n)
		if err != nil {
			continue
		}

		output += strconv.Itoa(n) + " | " +
			shortName(g.player1) + " | " + diceIcon(g.roll1) + " | " +
			shortName(g.player2) + " | " + diceIcon(g.roll2) + " | " +
			g.winner() + "\n"
	}

	// Display the leader board
	output += `
## Leader board:
Displays the current ranking of players based on their performance in Dice Roller

| Rank | Player                | Wins | Losses | Draws | Points |
|------|-----------------------|------|--------|-------|--------|
`

	for i, player := range getLeaderBoard() {
		rankIcon := strconv.Itoa(i)

		if i == 0 {
			rankIcon = "🥇"
		} else if i == 1 {
			rankIcon = "🥈"
		} else if i == 2 {
			rankIcon = "🥉"
		}

		output += ufmt.Sprintf("| %s   | %s | %d | %d | %d | %d |\n",
			rankIcon,
			shortName(player.addr),
			player.wins,
			player.losses,
			player.draws,
			player.points,
		)
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

// getGame retrieves the game state by its ID
func getGame(idx int) (*game, error) {
	v, ok := games.Get(seqid.ID(idx).String())
	if !ok {
		return nil, errors.New("game not found")
	}
	return v.(*game), nil
}

// updateResult updates the player's result and points based on the game outcome
func (p *player) updateResult(result int) {
	switch result {
	case win:
		p.wins++
		p.points += 3
	case lose:
		p.losses++
	case draw:
		p.draws++
		p.points++
	}
}

// getPlayer retrieves a player or initializes a new one if they don't exist
func getPlayer(addr std.Address) *player {
	v, ok := players.Get(addr.String())
	if !ok {
		player := &player{
			addr: addr,
		}
		players.Set(addr.String(), player)
		return player
	}

	return v.(*player)
}

// getLeaderBoard generates a leaderboard sorted by points
func getLeaderBoard() leaderBoard {
	board := leaderBoard{}
	players.Iterate("", "", func(key string, value interface{}) bool {
		player := value.(*player)
		board = append(board, *player)
		return false
	})

	sort.Sort(board)

	return board
}

// Methods for sorting the leaderboard
func (r leaderBoard) Len() int {
	return len(r)
}

func (r leaderBoard) Less(i, j int) bool {
	return r[i].points > r[j].points
}

func (r leaderBoard) Swap(i, j int) {
	r[i].points, r[j].points = r[j].points, r[i].points
}
