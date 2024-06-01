package flippandoserver

import (
	
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
func CreateGameBoard(tileType string, boardSize int, seed int64) ([]int64, []int64, int) {
	// Initialize the board and gameTiles slices based on boardSize
	board := make([]int64, boardSize)

	// Initialize the gameTiles slice with a length that is the square root of boardSize
	gameTilesLength := int(sqrt(float64(boardSize)))
	gameTiles := make([]int64, gameTilesLength)
	// @todo make call to generateRandomNumbers, so you can create the actual gameTiles array,
	rng := int64(4)
	if tileType == "squareGrid" {
		rng = int64(16)
	} else if tileType == "greyGradient" || tileType == "redGradient" || tileType == "blueGradient" || tileType == "greenGradient" {
		rng = int64(8)
	} else if tileType == "hexagrams" {
		rng = int64(4)
		// adjust for bigger board
		if len(gameTiles) == 8 {
			gameTiles = make([]int64, 4)
		}
	} else if tileType == "dice" {
		rng = int64(6)
		// adjust for bigger board
		if len(gameTiles) == 8 {
			gameTiles = make([]int64, 6)
		}
	}
	// Generating random numbers for gameTiles
    generatedNumbers := GenerateRandomNumbers(len(gameTiles), 1, int(rng), seed)
    for i, num := range generatedNumbers {
        gameTiles[i] = int64(num)
    }

	// we're returning discrete value, and do the json marshaling in realm
	return board, gameTiles, boardSize;
}

func FlipTiles(board []int64, solvedBoard []int64, gameTiles []int64, positions []int64, seed int64) ([]int64, []int64) {
	//   - returns the current game board, and the solved game board
	
	randomNumberSlice := GenerateRandomNumbers(2, 1, len(gameTiles), seed)

	// Safety check for 'positions' slice
    if len(positions) < 2 {
        return nil, nil  
    }

    // Safety check for 'randomNumberSlice'
    if randomNumberSlice == nil || len(randomNumberSlice) < 2 {
        return nil, nil  
    }
	
	// position hasn't been discovered before
	if board[positions[0]] == 0 {
		board[positions[0]] = gameTiles[randomNumberSlice[0] - 1]
	}

	if board[positions[1]] == 0 {
		board[positions[1]] = gameTiles[randomNumberSlice[1] -1]
	}

	// values are equal, update solved board
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

}

func GenerateRandomNumbers(howMany, min, max int, seed int64) []int {
	if min > max || howMany <= 0 || (max-min+1) < howMany {
		return nil
	}

	r := New(seed)
	numbers := make([]int, 0, howMany)
	seen := make(map[int]bool)

	for len(numbers) < howMany {
		randomNumber := r.Intn(max-min+1) + min

		// If the number hasn't been generated before, add it to the slice and the map
		if !seen[randomNumber] {
			seen[randomNumber] = true
			numbers = append(numbers, randomNumber)
		}
	}

	return numbers
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
