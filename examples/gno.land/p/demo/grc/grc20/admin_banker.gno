package grc20

import "std"

func AdminBanker(b *Bank) *adminBanker {
	prev := std.PrevRealm().Addr()
	if b.owner != prev {
		panic("AdminBanker() can only be called by the Bank owner")
	}
	return &adminBanker{
		bank: b,
	}
}
