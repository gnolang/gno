package game2048

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

	if b.IsWin() {
		b.Status = BoardStatusWin
		return
	}

	AddRandomTile()
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

	if b.IsWin() {
		b.Status = BoardStatusWin
		return
	}

	AddRandomTile()
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

	if b.IsWin() {
		b.Status = BoardStatusWin
		return
	}

	AddRandomTile()
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

	if b.IsWin() {
		b.Status = BoardStatusWin
		return
	}

	AddRandomTile()
}

func (b *Board) IsGameOver() bool {
	return b.Status == BoardStatusGameOver
}

func (b *Board) IsPlaying() bool {
	return b.Status == BoardStatusPlaying
}

func (b *Board) IsWin() bool {
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if b.Grid[i][j] == 2048 {
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
		b.Status = BoardStatusOver
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
			if b.Gridp[i][j] == 0 {
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
