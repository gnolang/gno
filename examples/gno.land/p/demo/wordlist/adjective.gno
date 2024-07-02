package wordlist

import "math/rand"

var (
	adjectives = []string{
		"angry",
		"brave",
		"cocky",
		"goofy",
		"happy",
		"jolly",
		"lucid",
		"nifty",
		"sharp",
		"stoic",
	}
	adjectivesL = len(adjectives)
)

func RandomAdjective(r *rand.Rand) string {
	return adjectives[r.IntN(adjectivesL)]
}
