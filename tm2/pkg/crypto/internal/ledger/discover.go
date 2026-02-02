// Package ledger contains the internals for package crypto/keys/ledger,
// primarily existing so that the Discover function can be mocked elsewhere.
package ledger

import (
	"errors"

	"github.com/btcsuite/btcd/btcec/v2"
	ledger_go "github.com/cosmos/ledger-cosmos-go"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/zondax/hid"
)

// SECP256K1 reflects an interface a Ledger API must implement for SECP256K1
type SECP256K1 interface {
	Close() error
	// Returns an uncompressed pubkey
	GetPublicKeySECP256K1([]uint32) ([]byte, error)
	// Returns a compressed pubkey and bech32 address (requires user confirmation)
	GetAddressPubKeySECP256K1([]uint32, string) ([]byte, string, error)
	// Signs a message (requires user confirmation)
	SignSECP256K1([]uint32, []byte, byte) ([]byte, error)
}

// Discover defines a function to be invoked at runtime for discovering
// a connected Ledger device.
var Discover DiscoverFn = DiscoverDefault

// DiscoverDefault is the default function for [Discover].
func DiscoverDefault() (SECP256K1, error) {
	if !hid.Supported() {
		return nil, errors.New("ledger support is not enabled, try building with CGO_ENABLED=1")
	}

	device, err := ledger_go.FindLedgerCosmosUserApp()
	if err != nil {
		return nil, err
	}

	return device, nil
}

// DiscoverMock can be used as a mock [DiscoverFn].
func DiscoverMock() (SECP256K1, error) {
	privateKey := secp256k1.GenPrivKey()

	_, pubKeyObject := btcec.PrivKeyFromBytes(privateKey[:])
	return discoverMock{
		pubKey:  pubKeyObject.SerializeCompressed(),
		address: privateKey.PubKey().Address().String(),
	}, nil
}

type discoverMock struct {
	MockLedger
	pubKey  []byte
	address string
}

func (m discoverMock) GetAddressPubKeySECP256K1(data []uint32, str string) ([]byte, string, error) {
	return m.pubKey, m.address, nil
}

// MockLedger is an interface that can be used to create mock [Ledger].
// Embed it in another type, and implement the method you want to mock:
//
//	type MyMock struct { MockLedger }
//	func (MyMock) SignSECP256K1(d1, d2 []byte, d3 byte) ([]byte, error) { ... }
type MockLedger struct{}

func (MockLedger) Close() error {
	return nil
}

func (MockLedger) GetPublicKeySECP256K1(data []uint32) ([]byte, error) {
	return nil, nil
}

func (MockLedger) GetAddressPubKeySECP256K1(data []uint32, str string) ([]byte, string, error) {
	return nil, "", nil
}

func (MockLedger) SignSECP256K1(d1 []uint32, d2 []byte, d3 byte) ([]byte, error) {
	return nil, nil
}

// DiscoverFn defines a Ledger discovery function that returns a
// connected device or an error upon failure. Its allows a method to avoid CGO
// dependencies when Ledger support is potentially not enabled.
type DiscoverFn func() (SECP256K1, error)
