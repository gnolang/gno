package leaderboard

import "strconv"

var scores = map[string]int{
	"alice": 10,
	"bob":   7,
	"carol": 12,
}

func Render(_ string) string {
	out := "# Scores\n\n"
	for name, score := range scores {
		out += "- " + name + ": " + strconv.Itoa(score) + "\n"
	}
	return out
}
