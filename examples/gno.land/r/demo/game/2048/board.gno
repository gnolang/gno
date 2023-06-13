package game2048

import (
	"std"

	"gno.land/p/demo/rand"
)

type BoardStatus int

const (
	BoardStatusPlaying = iota
	BoardStatusWin
	BoardStatusGameOver
)

const (
	BoardSize = 4
)

type Position struct {
	X int
	Y int
}

type Board struct {
	Grid   [][]int
	Status BoardStatus
}

func newBoard() *Board {
	grid := make([][]int, BoardSize)
	for i := 0; i < BoardSize; i++ {
		grid[i] = make([]int, BoardSize)
	}

	return &Board{
		Grid:   grid,
		Status: BoardStatusPlaying,
	}
}

func (b *Board) MoveLeft() {
	if b.Status != BoardStatusPlaying {
		panic("game is finished; reset to play again")
	}

	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if b.Grid[i][j] == 0 {
				continue
			}

			for k := j + 1; k < BoardSize; k++ {
				if b.Grid[i][k] == 0 {
					continue
				}

				if b.Grid[i][j] == b.Grid[i][k] {
					b.Grid[i][j] *= 2
					b.Grid[i][k] = 0
				}

				break
			}
		}
	}

	if b.SetIfWin() {
		return
	}
	if b.SetIfGameOver() {
		return
	}

	b.AddRandomTile()
}

func (b *Board) MoveRight() {
	if b.Status != BoardStatusPlaying {
		panic("game is finished; reset to play again")
	}

	for i := 0; i < BoardSize; i++ {
		for j := BoardSize - 1; j >= 0; j-- {
			if b.Grid[i][j] == 0 {
				continue
			}

			for k := j - 1; k >= 0; k-- {
				if b.Grid[i][k] == 0 {
					continue
				}

				if b.Grid[i][j] == b.Grid[i][k] {
					b.Grid[i][j] *= 2
					b.Grid[i][k] = 0
				}

				break
			}
		}
	}

	if b.SetIfWin() {
		return
	}
	if b.SetIfGameOver() {
		return
	}

	b.AddRandomTile()
}

func (b *Board) MoveUp() {
	if b.Status != BoardStatusPlaying {
		panic("game is finished; reset to play again")
	}

	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if b.Grid[i][j] == 0 {
				continue
			}

			for k := i + 1; k < BoardSize; k++ {
				if b.Grid[k][j] == 0 {
					continue
				}

				if b.Grid[i][j] == b.Grid[k][j] {
					b.Grid[i][j] *= 2
					b.Grid[k][j] = 0
				}

				break
			}
		}
	}

	if b.SetIfWin() {
		return
	}
	if b.SetIfGameOver() {
		return
	}

	b.AddRandomTile()
}

func (b *Board) MoveDown() {
	if b.Status != BoardStatusPlaying {
		panic("game is finished; reset to play again")
	}

	for i := 0; i < BoardSize; i++ {
		for j := BoardSize - 1; j >= 0; j-- {
			if b.Grid[i][j] == 0 {
				continue
			}

			for k := i + 1; k < BoardSize; k++ {
				if b.Grid[k][j] == 0 {
					continue
				}

				if b.Grid[i][j] == b.Grid[k][j] {
					b.Grid[i][j] *= 2
					b.Grid[k][j] = 0
				}

				break
			}
		}
	}

	if b.SetIfWin() {
		return
	}
	if b.SetIfGameOver() {
		return
	}

	b.AddRandomTile()
}

// SetIfGameOver checks if it is a game over and sets the board
// status to game over
func (b *Board) SetIfGameOver() bool {
	// Check empty tiles
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if b.Grid[i][j] == 0 {
				return false
			}
		}
	}

	// Check merging adjacent tiles horizontally
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize-1; j++ {
			if b.Grid[i][j] == b.Grid[i][j+1] {
				return false
			}
		}
	}

	// Check merging adjacent tiles vertically
	for i := 0; i < BoardSize-1; i++ {
		for j := 0; j < BoardSize; j++ {
			if b.Grid[i][j] == b.Grid[i+1][j] {
				return false
			}
		}
	}

	b.Status = BoardStatusGameOver
	return true
}

func (b *Board) IsPlaying() bool {
	return b.Status == BoardStatusPlaying
}

func (b *Board) IsWin() bool {
	return b.Status == BoardStatusWin
}

func (b *Board) IsGameOver() bool {
	return b.Status == BoardStatusGameOver
}

// SetIfWin checks if it is a win and sets the board status to win
func (b *Board) SetIfWin() bool {
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if b.Grid[i][j] == 2048 {
				b.Status = BoardStatusWin
				return true
			}
		}
	}

	return false
}

func (b *Board) Reset() {
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			b.Grid[i][j] = 0
		}
	}

	b.Status = BoardStatusPlaying
	b.AddRandomTile()
	b.AddRandomTile()
}

func (b *Board) AddRandomTile() {
	seed := std.GetHeight()
	val := []int{2, 4}

	emptyTiles := b.GetEmptyTiles()
	if len(emptyTiles) == 0 {
		b.Status = BoardStatusGameOver
	}

	r := rand.FromSeed(seed)
	pos := emptyTiles[r.Intn(len(emptyTiles))]
	randVal := val[r.Intn(len(val))]

	b.Grid[pos.X][pos.Y] = randVal
}

func (b *Board) GetEmptyTiles() []Position {
	pos := make([]Position, 0)
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if b.Grid[i][j] == 0 {
				pos = append(pos, Position{i, j})
			}
		}
	}

	return pos
}

func (b *Board) CalculateScore() int {
	score := 0
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			switch b.Grid[i][j] {
			case 2048:
				score += 2500
			case 1024:
				score += 500
			case 512:
				score += 250
			case 256:
				score += 125
			case 128:
				score += 75
			case 64:
				score += 50
			case 32:
				score += 25
			case 16:
				score += 10
			case 8:
				score += 5
			}
		}
	}

	return score
}
