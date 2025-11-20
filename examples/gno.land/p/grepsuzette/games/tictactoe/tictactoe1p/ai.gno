package tictactoe1p

import (
	"errors"

	"gno.land/p/grepsuzette/games/tictactoe"
)

// the default move picker.
func averageDifficulty(game *tictactoe.Game, Intn func(int) int) (x, y int, e error) {
	// candidates
	a := make([]struct{ x, y int }, 0)
	side := rune(1 + game.TurnNumber()%2)
	for y := 0; y <= 2; y++ {
		for x := 0; x <= 2; x++ {
			if game.At(x, y) != rune(0) {
				continue
			}
			// if can wins (and can see it), then win
			if game.WouldWin(side, x, y) && Intn(5) > 0 {
				return x, y, nil
			}
			a = append(a, struct{ x, y int }{x, y})
		}
	}
	if len(a) == 0 {
		return -1, -1, errors.New("no free tile left")
	} else {
		// random pick among candidates
		c := a[Intn(len(a))]
		return c.x, c.y, nil
	}
}
