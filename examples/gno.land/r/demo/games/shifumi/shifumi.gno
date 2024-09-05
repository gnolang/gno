package shifumi

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/r/demo/users"
)

const (
	empty = iota
	rock
	paper
	scissors
	last
)

type game struct {
	player1, player2 std.Address // shifumi is a 2-player game
	commit1, commit2 string      // commit hash for each player's move
	move1, move2     int         // actual move, revealed later
	salt1, salt2     string      // salts used for commit-reveal
}

var (
	games avl.Tree
	id    seqid.ID
)

// hashMove computes the SHA-256 hash of the given move and salt.
// This function is intended for on-chain use. To ensure consistency,
// users must implement an off-chain hashing function with the same logic.
// This means the off-chain implementation should use the same hash algorithm
// and concatenation of move and salt as shown here to generate the hash value
func hashMove(move int, salt string) string {
	hash := sha256.Sum256([]byte(strconv.Itoa(move) + salt))
	return hex.EncodeToString(hash[:])
}

// Commit allows a player to commit to a move (as a hash) without revealing it
func (g *game) commit(player std.Address, commit string) error {
	if player != g.player1 && player != g.player2 {
		return errors.New("invalid player")
	}
	if player == g.player1 && g.commit1 == "" {
		g.commit1 = commit
		return nil
	}
	if player == g.player2 && g.commit2 == "" {
		g.commit2 = commit
		return nil
	}
	return errors.New("already committed")
}

// Reveal allows the player to reveal their move and validate against the commit
func (g *game) reveal(player std.Address, move int, salt string) error {
	if !(move > empty && move < last) {
		return errors.New("invalid move")
	}
	commitHash := hashMove(move, salt)

	if player == g.player1 && g.move1 == empty && g.commit1 == commitHash {
		g.move1 = move
		g.salt1 = salt
		return nil
	}
	if player == g.player2 && g.move2 == empty && g.commit2 == commitHash {
		g.move2 = move
		g.salt2 = salt
		return nil
	}

	return errors.New("invalid reveal or already revealed")
}

func (g *game) winner() int {
	if g.move1 == empty || g.move2 == empty {
		return -1
	}
	if g.move1 == g.move2 {
		return 0
	}
	if g.move1 == rock && g.move2 == scissors ||
		g.move1 == paper && g.move2 == rock ||
		g.move1 == scissors && g.move2 == paper {
		return 1
	}
	return 2
}

// NewGame creates a new game where player1 is the caller and player2 the argument
func NewGame(player std.Address) int {
	games.Set(id.Next().String(), &game{player1: std.PrevRealm().Addr(), player2: player})
	return int(id)
}

// CommitMove allows a player to commit to a move using a hashed value.
// To use this function, the player must first generate a hash of their move
// and a random salt off-chain, then pass the hash as the commit parameter.
// Example call:
//
//	CommitMove(gameIndex, hashMove(move, salt))
//
// Where gameIndex is the ID of the game, and commit is the hashed move and salt.
func CommitMove(idx int, commit string) {
	v, ok := games.Get(seqid.ID(idx).String())
	if !ok {
		panic("game not found")
	}
	if err := v.(*game).commit(std.PrevRealm().Addr(), commit); err != nil {
		panic(err)
	}
}

// RevealMove allows a player to reveal their move and salt.
// To use this function, the player must provide their original move and salt used
// during the commit phase. The function will verify if the move and salt match the
// previously committed hash.
// Example call:
//
//	RevealMove(gameIndex, move, salt)
//
// Where gameIndex is the ID of the game, move is the player's move (1=rock, 2=paper, 3=scissors),
// and salt is the random string used during commit.
func RevealMove(idx, move int, salt string) {
	v, ok := games.Get(seqid.ID(idx).String())
	if !ok {
		panic("game not found")
	}

	if err := v.(*game).reveal(std.PrevRealm().Addr(), move, salt); err != nil {
		panic(err)
	}
}

func Render(path string) string {
	mov1 := []string{"", " 🤜  ", " ✋  ", " 👉  "}
	mov2 := []string{"", " 🤛  ", " ✋  ", " 👈  "}
	win := []string{"pending", "draw", "player1", "player2"}

	output := `# 👊 ✋ 👉 Shifumi
## Game Rules:
Shifumi is a two-player game (rock-paper-scissors). To play:
1. **Create a game** with an opponent using the [NewGame](shifumi?help&__func=NewGame) command
2. **Commit your move** as a hashed value (move + salt) using [CommitMove](shifumi?help&__func=CommitMove) with the game index and the hash
   - The **salt** is a randomly generated secret value that you should keep confidential. It ensures the commitment is secure
   - You can [compute your commit hash here](https://go.dev/play/p/ueRVMwFOFU1) by entering your move and salt
3. **Reveal your move** with the move number and the salt using [RevealMove](shifumi?help&__func=RevealMove)
   - Move values: 1 = Rock 🤜, 2 = Paper ✋, 3 = Scissors 👉
## Actions:
* [NewGame](shifumi?help&__func=NewGame) opponentAddress
* [CommitMove](shifumi?help&__func=CommitMove) gameIndex commitHash
* [RevealMove](shifumi?help&__func=RevealMove) gameIndex move salt (1=rock, 2=paper, 3=scissors)
## Recent Games:
 game  | player1 |     | player2 |       | win 
 ---   | ---      | --- | ---      | ---   | ---
`
	// Output the 100 most recent
	maxGames := 100
	for n := int(id); n > 0 && int(id)-n < maxGames; n-- {
		v, ok := games.Get(seqid.ID(n).String())
		if !ok {
			continue
		}
		g := v.(*game)
		output += strconv.Itoa(n) + " | " +
			shortName(g.player1) + " | " + mov1[g.move1] + " | " +
			shortName(g.player2) + " | " + mov2[g.move2] + " | " +
			win[g.winner()+1] + "\n"
	}
	return output
}

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
