package game2048

import (
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

func leaderboard() []*Player {
	// TODO: implement
	return nil
}
