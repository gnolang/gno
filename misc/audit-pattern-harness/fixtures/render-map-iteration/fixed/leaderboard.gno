package leaderboard

import (
	"strconv"

	"gno.land/p/nt/avl/v0"
)

var scores avl.Tree

func init() {
	// Fixed: tree iteration is sorted and deterministic for public Render output.
	scores.Set("alice", 10)
	scores.Set("bob", 7)
	scores.Set("carol", 12)
}

func Render(_ string) string {
	out := "# Scores\n\n"
	scores.Iterate("", "", func(name string, value any) bool {
		out += "- " + name + ": " + strconv.Itoa(value.(int)) + "\n"
		return false
	})
	return out
}
