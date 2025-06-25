package vls

import (
	"std"
	"gno.land/p/demo/avl"
)

// is uint64 enough?
// do we even keep the initial airdrop as part of this contract? it can be done in a separate contract or even a script

// address string -> amount int64
var InitialAirdropTree = avl.NewTree()

// init for now. think of a way to distribute governance tokens to the initial airdrop recipients
func Distribute() {
	InitialAirdropTree.Set("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y", int64(1)) // matijamarjanovic
	InitialAirdropTree.Set("g1...", int64(1)) //stefann-01
}

func InitialAirdrop() {
	Distribute()
	Airdrop(GovernanceAddr, InitialAirdropTree)
}