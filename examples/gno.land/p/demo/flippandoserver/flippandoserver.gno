package flippandoserver

import (
	"time"
	"std"
	"bytes"
	"strings"
	"strconv"

	"gno.land/p/demo/ufmt"
)

type GameBoard struct {
	Board     []int  `json:"board"`
	TileType  string `json:"tileType"`
	GameTiles []int  `json:"gameTiles"`
	BoardSize int    `json:"boardSize"`
}

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

func TestTiles(solvedBoard []int, size int) string {
	solvedSVG, err := GenerateColorSquareGrid(solvedBoard, size)
	if (err != nil){
		panic("error in GenerateColorSquareGrid")
	}
	return solvedSVG;
}

// creates game object and initializes board
func CreateGameBoard(tileType string, boardSize int) ([]int, []int, int) {
	// Initialize the board and gameTiles slices based on boardSize
	board := make([]int, boardSize)

	// Initialize the gameTiles slice with a length that is the square root of boardSize
	gameTilesLength := int(sqrt(float64(boardSize)))
	gameTiles := make([]int, gameTilesLength)
	// @todo make call to generateRandomNumbers, so you can create the actual gameTiles array,
	rng := 4
	if tileType == "squareGrid" || tileType == "hexagrams" {
		rng = 16
	} else if tileType == "dice" {
		rng := 6
	}
	gameTiles = generateRandomNumbers(4, rng)

	// we're returning discrete value, and do the json marshaling in realm
	return board, gameTiles, boardSize;
}

func FlipTiles(board []int, solvedBoard []int, gameTiles []int, positions []int) ([]int, []int) {
	//   - returns the current game board, and the solved game board
	
	randomNumberSlice := generateRandomNumbers(2, len(board))
	
	if board[positions[0]] == 0 {
		board[positions[0]] = gameTiles[randomNumberSlice[0]]
	}

	if board[positions[1]] == 0 {
		board[positions[1]] = gameTiles[randomNumberSlice[1]]
	}

	if board[positions[0]] == board[positions[1]] {
		solvedBoard[positions[0]] = board[positions[0]]
		solvedBoard[positions[1]] = board[positions[1]]
	}

	// enforce solvability
	quantumThreshold := int(sqrt(float64(len(board))))
	unsolvedTiles := 0

	for j := 0; j < len(board); j++ {
		if solvedBoard[j] == 0 {
			unsolvedTiles++
		}
	}

	if unsolvedTiles <= quantumThreshold {
		// replace the board with solvedBoard and redeploy
		if board[positions[0]] != board[positions[1]] {
			board[positions[0]] = board[positions[1]]
			solvedBoard[positions[0]] = board[positions[0]]
			solvedBoard[positions[1]] = board[positions[1]]
		}
	}

	return board, solvedBoard

	// check for game solved
	// todo: promote this check to the realm? that's where we send user events
	/*
	   unsolvedTiles = 0

	   for j := 0; j < len(board); j++ {
	       if(solvedBoard[j] == 0){
	           unsolvedTiles = unsolvedTiles + 1
	       }
	   }
	   if(unsolvedTiles == 0){
	       // call flippandoGameMaster finishGame
	       flippandoGameMaster.finishGame(id, games[id].player);
	       emit GameSolved(id, games[id])
	   }*/
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

/*
var randIntn = func(n int) int {
	return rand.Intn(n)
}*/

func generateRandomNumbers(x int, r int) []int {
	//time.Sleep(10 * time.Millisecond)
	now := time.Now().UnixNano()
	numbers := make([]int, x)

	for i := 0; i < x; i++ {
		now = (now*1664525 + 1013904223) % (1 << 32) // Linear congruential generator (LCG)
		numbers[i] = int(now % int64(r))
	}

	return numbers
}

/*
func generateRandomNumbers(num, start, end int) []int {
	result := make([]int, 0, 4)
	result = append(result, 2)
	result = append(result, 3)
	return result
}*/

func sqrt(x float64) float64 {
	if x == 0 || x == 1 {
		return x
	}

	// Start with an initial guess
	guess := x / 2.0
	prevGuess := 0.0

	// Use a small threshold to determine when to stop the approximation
	const threshold = 0.00001

	// Use math.Abs to calculate the absolute value
	abs := func(f float64) float64 {
		if f < 0 {
			return -f
		}
		return f
	}

	for abs(guess-prevGuess) > threshold {
		prevGuess = guess
		guess = 0.5 * (guess + x/guess)
	}

	return guess
}
