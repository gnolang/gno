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
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/users"
)

type game struct {
	player1, player2 *player
	roll1, roll2     int
	score1, score2   int
	round, maxRounds int
}

type player struct {
	addr                              std.Address
	wins, losses, draws, points, rank int
}

type leaderBoard []player

const (
	empty = iota
	win
	draw
	lose
)

var (
	games  avl.Tree
	gameId seqid.ID

	players  avl.Tree
	playerId seqid.ID

	seed = uint64(entropy.New().Seed())
)

// Roll a dice between 1 and 6
func rollDice() int {
	r := rand.New(rand.NewPCG(seed, 0xdeadbeef))

	return r.IntN(6) + 1
}

// NewGame initializes a new Dice Duel game with the specified number of rounds
func NewGame(addr std.Address, rounds int) int {
	if rounds <= 0 {
		panic("invalid number of rounds")
	}

	for i := 0; i <= 10; i++ {
		addr := testutils.TestAddress(ufmt.Sprintf("%d", i))
		players.Set(playerId.Next().String(), &player{
			addr:   addr,
			points: i + 1,
		})
	}

	games.Set(gameId.Next().String(), &game{
		player1:   &player{addr: std.PrevRealm().Addr()},
		player2:   &player{addr: addr},
		maxRounds: rounds,
	})
	return int(gameId)
}

// Play simulates a dice roll for the current player and updates the game state
func Play(idx int) int {
	g, err := getGame(idx)
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

	// Check if the game is over and update the player's stats
	if g.round >= g.maxRounds {
		if g.score1 > g.score2 {
			g.player1.updateResult(win)
			g.player2.updateResult(lose)
		} else if g.score2 > g.score1 {
			g.player2.updateResult(win)
			g.player1.updateResult(lose)
		} else {
			g.player1.updateResult(draw)
			g.player2.updateResult(draw)
		}
	}

	return roll
}

// play handles a player's dice roll and updates their score
func (g *game) play(player std.Address, roll int) error {
	if player != g.player1.addr && player != g.player2.addr {
		return errors.New("invalid player")
	}

	if g.round >= g.maxRounds {
		return errors.New("game over")
	}

	if player == g.player1.addr && g.roll1 == 0 {
		g.roll1 = roll
		g.score1 += roll
		return nil
	}

	if player == g.player2.addr && g.roll2 == 0 {
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

// Render provides a summary of the current state of all games and player rankings.
func Render(path string) string {
	// Create the output for the games
	output := `# 🎲 Dice Roller

## Game Rules:
Dice Roller is a two-player game where each player rolls a dice over a specified number of rounds
1. **Create a game** with an opponent using [NewGame](diceroller?help&__func=NewGame)
2. **Play rounds** using [Play](diceroller?help&__func=Play), where each player rolls a dice
   - Dice values range from 1 to 6 for each roll
   - Each player rolls the dice in turn, and the rolls are recorded for each round
## Actions:
* [NewGame](diceroller?help&__func=NewGame) opponentAddress rounds
  - **opponentAddress**: The address of the opponent player
  - **rounds**: The number of rounds for the game. Must be a positive integer
* [Play](diceroller?help&__func=Play) gameIndex
  - **gameIndex**: The unique index of the game you want to play
## Recent Games:
Displays a summary of the most recent games. Up to 10 recent games are shown

| Game | Player 1 | Roll 1 | Score 1 | Player 2 | Roll 2 | Score 2 | Round | Max Rounds | Winner |
|------|----------|--------|---------|----------|--------|---------|-------|------------|--------|
`

	// Display up to 10 recent games
	maxGames := 10
	for n := int(gameId); n > 0 && int(gameId)-n < maxGames; n-- {
		g, err := getGame(n)
		if err != nil {
			continue
		}

		output += strconv.Itoa(n) + " | " +
			shortName(g.player1.addr) + " | " + diceIcon(g.roll1) + " | " + strconv.Itoa(g.score1) + " | " +
			shortName(g.player2.addr) + " | " + diceIcon(g.roll2) + " | " + strconv.Itoa(g.score2) + " | " +
			strconv.Itoa(g.round) + " | " + strconv.Itoa(g.maxRounds) + " | " +
			g.winner() + "\n"
	}

	// Add ranking board to the output
	// Display the ranking board
	output += `
## Player Rankings:
Displays the current ranking of players based on their performance in Dice Roller.

| Rank | Player                | Wins | Losses | Draws | Points |
|------|-----------------------|------|--------|-------|--------|
`

	for i, player := range getRankingBoard() {
		// Determine the icon based on the rank
		rankIcon := strconv.Itoa(i)

		if i == 0 {
			rankIcon = "🥇"
		} else if i == 1 {
			rankIcon = "🥈"
		} else if i == 2 {
			rankIcon = "🥉"
		}

		// Format the player address to ensure it's aligned
		playerName := shortName(player.addr)
		if len(playerName) > 20 {
			playerName = playerName[:20] + "..."
		}

		// Add each player's rank information
		output += ufmt.Sprintf("| %s   | %s | %d | %d | %d | %d |\n",
			rankIcon,
			playerName,
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

func getPlayer(addr std.Address) *player {
	v, ok := players.Get(addr.String())
	if !ok {
		player := &player{
			addr: addr,
		}
		players.Set(playerId.Next().String(), player)
		return player
	}

	return v.(*player)
}

func getRankingBoard() leaderBoard {
	board := leaderBoard{}
	players.Iterate("", "", func(key string, value interface{}) bool {
		player := value.(*player)
		board = append(board, *player)
		return false
	})

	sort.Sort(board)

	return board
}

func (r leaderBoard) Len() int {
	return len(r)
}

func (r leaderBoard) Less(i, j int) bool {
	return r[i].points > r[j].points
}

func (r leaderBoard) Swap(i, j int) {
	r[i].points, r[j].points = r[j].points, r[i].points
}
