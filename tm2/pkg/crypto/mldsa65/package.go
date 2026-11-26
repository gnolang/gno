package mldsa65

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var Package = amino.RegisterPackage(amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/crypto/mldsa65",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	PubKeyMLDSA65{}, "PubKeyMLDSA65",
	PrivKeyMLDSA65{}, "PrivKeyMLDSA65",
))
