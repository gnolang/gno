package wordlist

import "math/rand"

var (
	colors = []string{
		"aqua",
		"black",
		"blue",
		"fuchsia",
		"gray",
		"green",
		"lime",
		"maroon",
		"navy",
		"olive",
		"purple",
		"silver",
		"teal",
		"white",
		"yellow",
	}
	colorsL = len(colors)
)

func RandomColor(r *rand.Rand) string {
	return colors[r.IntN(colorsL)]
}
