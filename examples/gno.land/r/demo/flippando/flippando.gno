package flippando

import (
	"fmt"
	"math"
)

type game struct {
	owner     string
	gameId    string
	board     []int
	tileType  string
	gameTiles []int
	boardSize int
	status    string
}

// owner - 'address', gameId - 'string', board - [int], tileType - 'coloredSquare', 'dice', 'hexagram', gameTiles - [int], boardSize - {4, 8}, status - 'created | ongoing | solved'

type positions struct {
	position1 int
	position2 int
}

// position1 - int, postion2 - int

type inTransit struct {
	tokenId   int
	inTransit bool
}

// tokenId - int, inTransit, bool

// creates game object and initializes board
func startGame(owner string, gameId string, tileType string, boardSize int) *game {
	// Initialize the board and gameTiles slices based on boardSize
	board := make([]int, boardSize)

	// Initialize the gameTiles slice with a length that is the square root of boardSize
	gameTilesLength := int(math.Sqrt(float64(boardSize)))
	gameTiles := make([]int, gameTilesLength)

	// Create a new game instance
	newGame := &game{
		owner:     owner,
		gameId:    gameId,
		board:     board,
		tileType:  tileType,
		gameTiles: gameTiles,
		boardSize: boardSize,
		status:    "initiated",
	}

	return newGame
}

func flipTiles(gameId string, board []int, solvedBoard []int, gameTiles []int, positions positions) {
	//   - returns the tiles at the positions[] in the board array
	//   - updates the game[gameId].board with the indices at positions
}

func createNFT(gameId string) {
	// - generates a GRC721 token with the solved board SVG as a tokenURI
}

func makeArt(owner string, buildingBlocks []int) {
	// - checks if any of the tokenIds in the buildingBlocks[] are belonging to the user (not allowed)
	// - checks if any of the tokenIds in the buildingBlocks[] are in transit
	// - unlocks and sends the locked Flips in each used NFT
	// - transfers the NFTs to the new owner (the art creator)
	// - calls FlippandoBundler.bundleAssets() to generate a new GRC721 NFT with a standard tokenURI, and the tokenIds of the used NFT primitives
}

func resetGame(gameId string) {
	//- utility function, nullifies the game[gameId]
}

func boardHasZeroValues(gameId string) {
	// - utility function, checks for existing zeros in the board (game is ongoing) // do we still need this?
}

// func generateRandomNumbers(num, start, end int) []int -> moved to package

func isValidGameLevel(boardSize int) {
	// - utility function, checks valid game levels
}

func isValidGameType(tileType string) {
	//   - utility function, checks valid tile types
}
