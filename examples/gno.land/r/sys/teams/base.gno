package teams

import (
	"std"

	"gno.land/p/demo/avl"
)

type BaseTeam struct {
	address std.Address
	members avl.Tree // std.Address -> struct{}{}
}

func (bt *BaseTeam) Has(member std.Address) bool {
	return bt.members.Has(member.String())
}

func (bt *BaseTeam) Can(member std.Address, do ActionVerb, on ...std.Address) bool {
	return true
}
