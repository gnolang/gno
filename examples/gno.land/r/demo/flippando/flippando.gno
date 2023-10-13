package flippando

import (
	"bytes"
	"errors"
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/flippandoserver"
)

type GameBoard struct {
	Board     []int  `json:"board"`
	TileType  string `json:"tileType"`
	GameTiles []int  `json:"gameTiles"`
	BoardSize int    `json:"boardSize"`
}

type Game struct {
	ID 			string `json:"id"`
	GameStatus 	string	`json:"gameStatus"`
	GameBoard 	string	`json:"gameBoard"`
}

type inTransit struct {
	tokenId   int
	inTransit bool
}

// tokenId - int, inTransit, bool

// realm state
var (	
	gameStore     avl.Tree // string (game ID) -> *Game
	gameIDCounter uint64

	user2Games avl.Tree // std.Address -> []*Game
)


// creates game object and initializes board
func StartGame(player std.Address, tileType string, boardSize int) string {
	
	games := getUserGames(player)
	// Set up Game struct. Save in gameStore and user2games.
	gameIDCounter++
	newBoardGameJSON := flippandoserver.CreateGameBoard(tileType, boardSize)
	// id is zero-padded to work well with avl's alphabetic order.
	id := zeroPad9(strconv.FormatUint(gameIDCounter, 10))
	g := &Game{
		ID:        id,
		GameBoard: newBoardGameJSON,
		GameStatus: "initialized",
	}

	gameStore.Set(g.ID, g)
	addToUser2Games(player, g)
	// write the marshalJSON function for game
	// call marshall JSON and return the string
	newGameJSON, err := g.MarshalJSON()
	if err != nil {
		return "error";
	}
	return string(newGameJSON)
}

func (g Game) MarshalJSON() (_ []byte, err error) {
	var b bytes.Buffer

	b.WriteByte('{')
	
	b.WriteString(`"id":"`)
	b.WriteString(g.ID)
	b.WriteString(`",`)

	b.WriteString(`"gameBoard":`)
	b.WriteString(g.GameBoard) 
	b.WriteString(`,`)
	
	b.WriteString(`"gameStatus":"`)
	b.WriteString(g.GameStatus)
	b.WriteString(`"`)
	
	b.WriteByte('}')

	return b.Bytes(), nil
}


func FlipTiles(gameId string, board []int, solvedBoard []int, gameTiles []int, positions []int) {
	//   - returns the tiles at the positions[] in the board array
	//   - updates the game[gameId].board with the indices at positions
	flippandoserver.FlipTiles(gameId, board, solvedBoard, gameTiles, positions)
}

func addToUser2Games(addr std.Address, game *Game) {
	var games []*Game
	v, ok := user2Games.Get(string(addr))
	if ok {
		games = v.([]*Game)
	}
	// game must be at top, because it is the latest ID
	games = append([]*Game{game}, games...)
	user2Games.Set(string(addr), games)
}

func getUserGames(user std.Address) []*Game {
	val, exist := user2Games.Get(user.String())
	var games []*Game
	if !exist {
		return nil
	}
	return val.([]*Game)
}

func zeroPad9(s string) string {
	n := 9 - len(s)
	if n < 0 {
		return s
	}
	return zeroes[:n] + s
}
const zeroes = "000000000"

func createNFT(gameId string) {
	// private function, called when the game ends
	// - generates a GRC721 token with the solved board SVG as a tokenURI
}

func makeArt(owner string, buildingBlocks []int) {
	// - checks if any of the tokenIds in the buildingBlocks[] are belonging to the user (not allowed)
	// - checks if any of the tokenIds in the buildingBlocks[] are in transit
	// - unlocks and sends the locked Flips in each used NFT
	// - transfers the NFTs to the new owner (the art creator)
	// - calls FlippandoBundler.bundleAssets() to generate a new GRC721 NFT with a standard tokenURI, and the tokenIds of the used NFT primitives
}

func ResetGame(gameId string) {
	//- utility function, nullifies the game[gameId]
}

// utility, private functions
func boardHasZeroValues(gameId string) {
	// - utility function, checks for existing zeros in the board (game is ongoing) // do we still need this?
}

func isValidGameLevel(boardSize int) {
	// - utility function, checks valid game levels
}

func isValidGameType(tileType string) {
	//   - utility function, checks valid tile types
}

func RenderHome() string {
	return "render path"
}

// utils, to move into a separate package?

func sqrt(x float64) float64 {
	if x == 0 || x == 1 {
		return x
	}

	// Start with an initial guess
	guess := x / 2.0
	prevGuess := 0.0

	// Use a small threshold to determine when to stop the approximation
	const threshold = 0.00001

	for abs(guess-prevGuess) > threshold {
		prevGuess = guess
		guess = 0.5 * (guess + x/guess)
	}

	return guess
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
