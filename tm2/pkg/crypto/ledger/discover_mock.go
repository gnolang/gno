//go:build ledger_suite
// +build ledger_suite

package ledger

import (
	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// discoverLedger defines a function to be invoked at runtime for discovering
// a connected Ledger device.
var discoverLedger discoverLedgerFn = func() (LedgerSECP256K1, error) {
	privateKey := secp256k1.GenPrivKey()

	_, pubKeyObject := btcec.PrivKeyFromBytes(privateKey[:])

	return &MockLedger{
		GetAddressPubKeySECP256K1Fn: func(data []uint32, str string) ([]byte, string, error) {
			return pubKeyObject.SerializeCompressed(), privateKey.PubKey().Address().String(), nil
		},
	}, nil
}

type (
	closeDelegate                     func() error
	getPublicKeySECP256K1Delegate     func([]uint32) ([]byte, error)
	getAddressPubKeySECP256K1Delegate func([]uint32, string) ([]byte, string, error)
	signSECP256K1Delegate             func([]uint32, []byte, byte) ([]byte, error)
)

type MockLedger struct {
	CloseFn                     closeDelegate
	GetPublicKeySECP256K1Fn     getPublicKeySECP256K1Delegate
	GetAddressPubKeySECP256K1Fn getAddressPubKeySECP256K1Delegate
	SignSECP256K1Fn             signSECP256K1Delegate
}

func (m *MockLedger) Close() error {
	if m.CloseFn != nil {
		return m.CloseFn()
	}

	return nil
}

func (m *MockLedger) GetPublicKeySECP256K1(data []uint32) ([]byte, error) {
	if m.GetPublicKeySECP256K1Fn != nil {
		return m.GetPublicKeySECP256K1Fn(data)
	}

	return nil, nil
}

func (m *MockLedger) GetAddressPubKeySECP256K1(data []uint32, str string) ([]byte, string, error) {
	if m.GetAddressPubKeySECP256K1Fn != nil {
		return m.GetAddressPubKeySECP256K1Fn(data, str)
	}

	return nil, "", nil
}

func (m *MockLedger) SignSECP256K1(d1 []uint32, d2 []byte, d3 byte) ([]byte, error) {
	if m.SignSECP256K1Fn != nil {
		return m.SignSECP256K1Fn(d1, d2, d3)
	}

	return nil, nil
}
