package dice_roller

import (
	"errors"
	"math/rand"
	"sort"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/entropy"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"

	"gno.land/r/sys/users"
)

type (
	// game represents a Dice Roller game between two players
	game struct {
		player1, player2 address
		roll1, roll2     int
	}

	// player holds the information about each player including their stats
	player struct {
		addr                        address
		wins, losses, draws, points int
	}

	// leaderBoard is a slice of players, used to sort players by rank
	leaderBoard []player
)

const (
	// Constants to represent game result outcomes
	ongoing = iota
	win
	draw
	loss
)

var (
	games  avl.Tree // AVL tree for storing game states
	gameId seqid.ID // Sequence ID for games

	players avl.Tree // AVL tree for storing player data

	seed = uint64(entropy.New().Seed())
	r    = rand.New(rand.NewPCG(seed, 0xdeadbeef))
)

// rollDice generates a random dice roll between 1 and 6
func rollDice() int {
	return r.IntN(6) + 1
}

// NewGame initializes a new game with the provided opponent's address
func NewGame(addr address) int {
	if !addr.IsValid() {
		panic("invalid opponent's address")
	}

	games.Set(gameId.Next().String(), &game{
		player1: std.PreviousRealm().Address(),
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
	if err := g.play(std.PreviousRealm().Address(), roll); err != nil {
		panic(err)
	}

	// If both players have rolled, update the results and leaderboard
	if g.isFinished() {
		// If the player is playing against themselves, no points are awarded
		if g.player1 == g.player2 {
			return roll
		}

		player1 := getPlayer(g.player1)
		player2 := getPlayer(g.player2)

		if g.roll1 > g.roll2 {
			player1.updateStats(win)
			player2.updateStats(loss)
		} else if g.roll2 > g.roll1 {
			player2.updateStats(win)
			player1.updateStats(loss)
		} else {
			player1.updateStats(draw)
			player2.updateStats(draw)
		}
	}

	return roll
}

// play processes a player's roll and updates their score
func (g *game) play(player address, roll int) error {
	if player != g.player1 && player != g.player2 {
		return errors.New("invalid player")
	}

	if g.isFinished() {
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

// isFinished checks if the game has ended
func (g *game) isFinished() bool {
	return g.roll1 != 0 && g.roll2 != 0
}

// checkResult returns the game status as a formatted string
func (g *game) status() string {
	if !g.isFinished() {
		return resultIcon(ongoing) + " Game still in progress"
	}

	if g.roll1 > g.roll2 {
		return resultIcon(win) + " Player1 Wins !"
	} else if g.roll2 > g.roll1 {
		return resultIcon(win) + " Player2 Wins !"
	} else {
		return resultIcon(draw) + " It's a Draw !"
	}
}

// Render provides a summary of the current state of games and leader board
func Render(path string) string {
	var sb strings.Builder

	sb.WriteString(`# 🎲 **Dice Roller Game**

Welcome to Dice Roller! Challenge your friends to a simple yet exciting dice rolling game. Roll the dice and see who gets the highest score !

---

## **How to Play**:
1. **Create a game**: Challenge an opponent using [NewGame](./dice_roller$help&func=NewGame)
2. **Roll the dice**: Play your turn by rolling a dice using [Play](./dice_roller$help&func=Play)

---

## **Scoring Rules**:
- **Win** 🏆: +3 points
- **Draw** 🤝: +1 point each
- **Lose** ❌: No points
- **Playing against yourself**: No points or stats changes for you

---

## **Recent Games**:
Below are the results from the most recent games. Up to 10 recent games are displayed

| Game | Player 1 | 🎲 Roll 1 | Player 2 | 🎲 Roll 2 | 🏆 Winner |
|------|----------|-----------|----------|-----------|-----------|
`)

	maxGames := 10
	for n := int(gameId); n > 0 && int(gameId)-n < maxGames; n-- {
		g, err := getGame(n)
		if err != nil {
			continue
		}

		sb.WriteString(strconv.Itoa(n) + " | " +
			"<span title=\"" + string(g.player1) + "\">" + shortName(g.player1) + "</span>" + " | " + diceIcon(g.roll1) + " | " +
			"<span title=\"" + string(g.player2) + "\">" + shortName(g.player2) + "</span>" + " | " + diceIcon(g.roll2) + " | " +
			g.status() + "\n")
	}

	sb.WriteString(`
---

## **Leaderboard**:
The top players are ranked by performance. Games played against oneself are not counted in the leaderboard

| Rank | Player                | Wins | Losses | Draws | Points |
|------|-----------------------|------|--------|-------|--------|
`)

	for i, player := range getLeaderBoard() {
		sb.WriteString(ufmt.Sprintf("| %s   | <span title=\""+string(player.addr)+"\">**%s**</span> | %d | %d | %d | %d |\n",
			rankIcon(i+1),
			shortName(player.addr),
			player.wins,
			player.losses,
			player.draws,
			player.points,
		))
	}

	sb.WriteString("\n---\n**Good luck and have fun !** 🎉")
	return sb.String()
}

// shortName returns a shortened name for the given address
func shortName(addr address) string {
	user := users.ResolveAddress(addr)
	if user != nil {
		return user.Name()
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

// updateResult updates the player's stats and points based on the game outcome
func (p *player) updateStats(result int) {
	switch result {
	case win:
		p.wins++
		p.points += 3
	case loss:
		p.losses++
	case draw:
		p.draws++
		p.points++
	}
}

// getPlayer retrieves a player or initializes a new one if they don't exist
func getPlayer(addr address) *player {
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
	players.Iterate("", "", func(key string, value any) bool {
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
	if r[i].points != r[j].points {
		return r[i].points > r[j].points
	}

	if r[i].wins != r[j].wins {
		return r[i].wins > r[j].wins
	}

	if r[i].draws != r[j].draws {
		return r[i].draws > r[j].draws
	}

	return false
}

func (r leaderBoard) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
