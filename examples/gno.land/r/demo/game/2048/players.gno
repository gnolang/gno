package game2048

import (
	"std"

	"gno.land/p/demo/avl"
)

var players avl.Tree // id -> *Player

func getPlayer(addr std.Address) *Player {
	player, found := players.Get(addr.String())
	if !found {
		return nil
	}

	return player.(*Player)
}

// Returns the top 10 players
func leaderboard() []*Player {
	res := make([]*Player, 0, 10)

	players.Iterate("", "", func(key string, value interface{}) bool {
		// TODO: sort by score and add only top 10
		res = append(res, value.(*Player))
		return false
	})

	return res
}
