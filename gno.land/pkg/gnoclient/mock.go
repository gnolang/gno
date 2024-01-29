package gnoclient

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type (
	mockSign     func(cfg SignCfg) (*std.Tx, error)
	mockInfo     func() keys.Info
	mockValidate func() error
)

type mockSigner struct {
	sign     mockSign
	info     mockInfo
	validate mockValidate
}

func (m *mockSigner) Sign(cfg SignCfg) (*std.Tx, error) {
	if m.sign != nil {
		return m.sign(cfg)
	}

	return nil, nil
}

func (m *mockSigner) Info() keys.Info {
	if m.info != nil {
		return m.info()
	}

	return nil
}

func (m *mockSigner) Validate() error {
	if m.validate != nil {
		return m.validate()
	}

	return nil
}

type mockKeysInfo struct{}

func (m mockKeysInfo) GetAddress() crypto.Address {
	adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	return adr
}

func (m mockKeysInfo) GetType() keys.KeyType {
	return 0
}

func (m mockKeysInfo) GetName() string {
	return "mockKeyInfoName"
}

func (m mockKeysInfo) GetPubKey() crypto.PubKey {
	pubkey, _ := crypto.PubKeyFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	return pubkey
}

func (m mockKeysInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, nil
}
