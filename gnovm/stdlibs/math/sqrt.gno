package math

import (
	"errors"
	"math"
)

func Sqrt(x float64, threshold float64) (float64, error) {
	if x < 0 {
		return 0, errors.New("input value must be non-negative")
	}
	if x == 0 || x == 1 {
		return x, nil
	}

	guess := x / 2.0
	prevGuess := 0.0

	for math.Abs(guess-prevGuess) > threshold {
		prevGuess = guess
		guess = 0.5 * (guess + x/guess)
	}

	return guess, nil
}
