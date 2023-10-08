package flippando

import (
	"bytes"
	"errors"

	"gno.land/p/demo/flippandoserver"
)

type inTransit struct {
	tokenId   int
	inTransit bool
}

// tokenId - int, inTransit, bool

// creates game object and initializes board
func StartGame(gameId string, tileType string, boardSize int) string {
	return flippandoserver.CreateGameBoard(gameId, tileType, boardSize)
}

func FlipTiles(gameId string, board []int, solvedBoard []int, gameTiles []int, positions []int) {
	//   - returns the tiles at the positions[] in the board array
	//   - updates the game[gameId].board with the indices at positions
	flippandoserver.FlipTiles(gameId, board, solvedBoard, gameTiles, positions)
}

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
