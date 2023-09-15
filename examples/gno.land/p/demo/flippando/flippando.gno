package flippando

import (
	"math/rand"
	"std"
	"time"

	"gno.land/p/demo/ufmt"
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

func flipTiles(board []int, solvedBoard []int, gameTiles []int, positions positions) ([]int, []int) {
	//   - returns the tiles at the positions[] in the board array
	//   - updates the game[gameId].board with the indices at positions

	// pos := positions
	// optimize for one call, and use the tuple?
	randomNumberSlice := generateRandomNumbers(2, 1, len(board))
	// randomNumber1 := generateRandomNumbers(1, 1, len(board))
	// randomNumber2 := generateRandomNumbers(1, 1, len(board))

	if board[positions.position1] == 0 {
		board[positions.position1] = gameTiles[randomNumberSlice[0]]
	}

	if board[positions.position2] == 0 {
		board[positions.position2] = gameTiles[randomNumberSlice[1]]
	}

	if board[positions.position1] == board[positions.position2] {
		solvedBoard[positions.position1] = board[positions.position1]
		solvedBoard[positions.position2] = board[positions.position2]
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
		if board[positions.position1] != board[positions.position2] {
			board[positions.position1] = board[positions.position2]
			solvedBoard[positions.position1] = board[positions.position1]
			solvedBoard[positions.position2] = board[positions.position2]
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

func generateRandomNumbers(num, start, end int) []int {
	if start > end {
		start, end = end, start // Ensure the range is in ascending order
	}

	if num < 0 {
		ufmt.Sprintf("The number of random values (num) must be non-negative")
		return nil
	}

	if num > end-start+1 {
		ufmt.Sprintf("The number of random values (num) cannot exceed the range size")
		return nil
	}

	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())
	// Generate num unique random values within the range
	result := make([]int, 0, num)
	for len(result) < num {
		randomValue := rand.Intn(end-start+1) + start

		// Check if the value already exists in the result slice
		// If not, append it to the result slice
		unique := true
		for _, val := range result {
			if val == randomValue {
				unique = false
				break
			}
		}

		if unique {
			result = append(result, randomValue)
		}
	}

	return result
}

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
