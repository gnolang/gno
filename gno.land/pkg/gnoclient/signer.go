package gnoclient

import "github.com/gnolang/gno/tm2/pkg/crypto/keys"

// Signer ...
type Signer interface {
	Sign()
}

type SignerFromKeybase struct {
	Keybase  keys.Keybase
	Account  string // name or bech32
	Password string // encryption password
}

var _ Signer = (*SignerFromKeybase)(nil)

func SignerFromKeybase(kb keys.Keybase, nameOrBech32 string, passwd string) Signer {

}

// InmemSignerFromBip39 creates an inmemory keybase which loads a single "default" account.
// It is intended to be used in systems that cannot rely on filesystem to store the private keys.
//
// Warning: It's recommended to use keys.NewKeyBaseFromDir when possible.
func InmemSignerFromBip39(mnemo string, passphrase string, account uint32, index uint32) (keys.Keybase, error) {
	kb := keys.NewInMemory()
	name := "default"
	passwd := "" // not needed in memory
	_, err := kb.CreateAccount(name, mnemo, passphrase, passwd, account, index)
	if err != nil {
		return nil, err
	}
	return kb, nil
}
