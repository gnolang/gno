package flippando

import (
	"strconv"
	"strings"
)

// used in marshaling structs with arrays
func writeIntSlice(slice []int64) string {
    var b strings.Builder
    for i, v := range slice {
        if i > 0 {
            b.WriteString(",")
        }
        b.WriteString(strconv.FormatInt(v, 10))  // converting int64 to string
    }
    return b.String()
}


// used in parsing params sent as JSON.stringified arrays of ints
func ParseString(s string) ([]int64, error) {
	s = strings.Trim(s, "[]")
	parts := strings.Split(s, ",")
	ints := make([]int64, len(parts))
	for i, part := range parts {
		val, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		ints[i] = int64(val)
	}
	return ints, nil
}

func zeroPad9(s string) string {
	n := 9 - len(s)
	if n < 0 {
		return s
	}
	return zeroes[:n] + s
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