package game2048

import (
	"std"
)

type Player struct {
	Id    std.Address
	Score int
	Board *Board
}

func newPlayer(id std.Address) *Player {
	return &Player{
		Id:    id,
		Score: 0,
		Board: newBoard(),
	}
}
