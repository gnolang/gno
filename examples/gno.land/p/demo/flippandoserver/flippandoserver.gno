package flippandoserver

import (
	"time"
	"math"
	"std"
	"bytes"
	"strings"
	"strconv"

	"gno.land/p/demo/ufmt"
)

type GameBoard struct {
	Board     []int64  `json:"board"`
	TileType  string `json:"tileType"`
	GameTiles []int64  `json:"gameTiles"`
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
/*
func TestTiles(solvedBoard []int64, size int) string {
	solvedSVG, err := GenerateColorSquareGrid(solvedBoard, size)
	if (err != nil){
		panic("error in GenerateColorSquareGrid")
	}
	return solvedSVG;
}*/

// creates game object and initializes board
func CreateGameBoard(tileType string, boardSize int) ([]int64, []int64, int) {
	// Initialize the board and gameTiles slices based on boardSize
	board := make([]int64, boardSize)

	// Initialize the gameTiles slice with a length that is the square root of boardSize
	gameTilesLength := int(sqrt(float64(boardSize)))
	gameTiles := make([]int64, gameTilesLength)
	// @todo make call to generateRandomNumbers, so you can create the actual gameTiles array,
	rng := int64(4)
	if tileType == "squareGrid" {
		rng = int64(16)
	} else if tileType == "hexagrams" {
		rng = int64(4)
	} else if tileType == "dice" {
		rng = int64(6)
	}
	gameTiles = generateRandomNumbers(4, int64(1), rng)
	// test
	//gameTiles := [1, 5, 3, 8]

	// we're returning discrete value, and do the json marshaling in realm
	return board, gameTiles, boardSize;
}

func FlipTiles(board []int64, solvedBoard []int64, gameTiles []int64, positions []int64) ([]int64, []int64) {
	//   - returns the current game board, and the solved game board
	
	randomNumberSlice := generateRandomNumbers(2, 1, int64(len(board)/4))
	
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

/*
func generateRandomNumbers(howMany int, min int64, max int64) []int64 {
	var rng = &rngSource{}
	rng.Seed(int64(time.Now()) + std.GetHeight())
    numbers := make([]int64, howMany)
    rangeSize := max - min + 1
    for i := 0; i < howMany; i++ {
        r := rng.Int63() 
        numbers[i] = min + (r % rangeSize)
    }
    return numbers
}*/

func generateRandomNumbers(howMany int, min int64, max int64) []int64 {
	var numbers []int64
	/*
	var rng = &rngSource{}
	rng.Seed(int64(time.Now()) + std.GetHeight(),)
	l := int64(0)
	rangeSize := max - min + 1
	for i := 0; i < howMany; i++ {
		r := rng.Int63() % l
		numbers[i] = min + (r % rangeSize)
	}*/
	numbers = []int64{1, 2, 3, 4}
	return numbers
}
/*
func generateRandomNumbers(howMany int, min, max int64) []int64 {
    // Mockup: just returning a slice of fixed values
    numbers := []int64{1, 2, 3, 4}  // Adjust these values as needed
    return numbers
}*/

/*func generateRandomNumbers(x int, r int) []int64 {
	//time.Sleep(10 * time.Millisecond)
	now := time.Now().UnixNano()
	numbers := make([]int64, x)

	for i := 0; i < x; i++ {
		height := std.GetHeight()
		now = (now*1664525 + 1013904223) % (1 << 32) + (height*255) // Linear congruential generator (LCG)
		//numbers[i] = int(now % int64(r))
		numbers[i] = int(int64(now) % int64(r))
		//numbers[i] = int(math.Abs(float64(now)) % float64(r))

	}

	return numbers
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
