package valopers

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

const (
	testEd25519PubKey   = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqkn3nln2k8glpf9kpqtz4h63n7cd0jvxkq6rpk4dxzp4sq527mhqwv40hr"
	testSecp256k1PubKey = "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj"
)

func TestPubKeyTypeURL(t *testing.T) {
	got, err := pubKeyTypeURL(testEd25519PubKey)
	uassert.NoError(t, err)
	uassert.Equal(t, "/tm.PubKeyEd25519", got)

	got, err = pubKeyTypeURL(testSecp256k1PubKey)
	uassert.NoError(t, err)
	uassert.Equal(t, "/tm.PubKeySecp256k1", got)

	_, err = pubKeyTypeURL("not-a-valid-bech32-pubkey")
	uassert.Error(t, err)
}

func TestAssertPubKeyTypeAllowed(t *testing.T) {
	// Empty allow-list accepts any well-formed type.
	testing.SetSysParamStrings("node", "valset", "pubkey_types", []string{})
	uassert.NotPanics(t, cur, func() { assertPubKeyTypeAllowed(testSecp256k1PubKey) })

	// ed25519-only allow-list: ed25519 accepted, secp256k1 rejected.
	testing.SetSysParamStrings("node", "valset", "pubkey_types", []string{"/tm.PubKeyEd25519"})
	uassert.NotPanics(t, cur, func() { assertPubKeyTypeAllowed(testEd25519PubKey) })
	uassert.PanicsWithMessage(t, cur, ErrDisallowedPubKeyType.Error(), func() {
		assertPubKeyTypeAllowed(testSecp256k1PubKey)
	})

	// Reset so the non-empty allow-list doesn't leak into other tests.
	testing.SetSysParamStrings("node", "valset", "pubkey_types", []string{})
}
