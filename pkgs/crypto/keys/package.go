package keys

import (
	"github.com/gnolang/gno/pkgs/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/pkgs/crypto/keys",
	"tm.keys",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	localInfo{}, "LocalInfo",
	ledgerInfo{}, "LedgerInfo",
	offlineInfo{}, "OfflineInfo",
	multiInfo{}, "MultiInfo",
))
