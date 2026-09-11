package mock

import (
	"sync"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

// Package is deliberately not registered with the global amino codec here, the
// way every other crypto package registers itself. Use
// UnsafeRegisterAminoPackage, and read its docs first.
var Package = amino.NewPackage(
	"github.com/gnolang/gno/tm2/pkg/crypto/mock",
	"tm",
	amino.GetCallersDirname(),
).WithDependencies().WithTypes(
	PubKeyMock{}, "PubKeyMock",
	PrivKeyMock{}, "PrivKeyMock",
)

var registerOnce sync.Once

// UnsafeRegisterAminoPackage registers PubKeyMock and PrivKeyMock with the
// global amino codec, so that they can be encoded to and decoded from amino
// bytes anywhere in the process. It is idempotent.
//
// Do NOT call this on any production path. These types stand in for real
// crypto, and PubKeyMock in particular verifies any signature that can be
// derived from the message and the public key alone — which is to say, by
// anyone. Registering it lets an caller-chosen type URL in untrusted bytes
// resolve to PubKeyMock, so any crypto.PubKey decoded from the wire may hold a
// key whose signatures are forgeable. That reaches further than it first looks:
// a key type needs only to be accepted somewhere a crypto.PubKey is decoded,
// such as a constituent key of a multisig.PubKeyMultisigThreshold, for its
// forgeable signatures to stand in for a real one.
//
// Registration is kept out of this package's init so that importing the package
// cannot enable that decoding as a side effect, and so that enabling it is
// greppable. Call it from an init() in a _test.go file: the global codec seals
// itself the first time it is used, and registering after that panics with
// "codec sealed".
func UnsafeRegisterAminoPackage() {
	registerOnce.Do(func() { amino.RegisterPackage(Package) })
}
