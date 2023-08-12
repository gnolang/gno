package gnoverse

import "github.com/gnolang/gno/tm2/pkg/std"

type account struct{ std.BaseAccount }

func protoAccount() std.Account { return &account{} }
