package dev

import (
	"fmt"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// type Keybase struct {
// 	keys.Keybase
// }

// func NewKeybase() *Keybase {
// 	return &Keybase{
// 		Keybase: keys.NewInMemory(),
// 	}
// }

// func (to *Keybase) ImportKeybaseFromPath(path string) error {
// 	from, err := keys.NewKeyBaseFromDir(path)
// 	if err != nil {
// 		return fmt.Errorf("unable to load keybase: %w", err)
// 	}

// 	keys, err := from.List()
// 	if err != nil {
// 		return fmt.Errorf("unable to list keys: %w", path, err)
// 	}

// 	for _, key := range keys {
// 		armor, err := from.Export(key.GetName())
// 		if err != nil {
// 			return fmt.Errorf("unable to import key %q: %w", key.GetName(), err)
// 		}

// 		err = to.Import(key.GetName(), armor)
// 		if err != nil {
// 			return fmt.Errorf("unable to import key %q: %w", key.GetName(), err)
// 		}
// 	}

// 	return nil
// }

// type PackagePath struct {
// 	Path                 string
// 	CreatorNameOrAddress string
// }

// func ParsePackagePath(path string) (PackagePath, error) {
// 	var ppath PackagePath

// 	upath, err := url.Parse(path)
// 	if err != nil {
// 		return ppath, fmt.Errorf("unable to parse package path: %w", err)
// 	}

// 	// Get path
// 	ppath.Path = filepath.Clean(upath.Path)
// 	// Check for options
// 	ppath.CreatorNameOrAddress = upath.Query().Get("creator")
// 	return ppath, nil
// }

// func LoadKeyabaseBalanceFromPath(kb keys.Keybase) ([]gnoland.Balance, error) {
// 	keys, err := kb.List()
// 	if err != nil {
// 		return nil, nil
// 	}

// 	for _, info := range keys {
// 		info.GetName()
// 	}
// }

// loadAccount with the given name and adds it to the keybase.
func loadAccount(kb keys.Keybase, accountName string) (gnoland.Balance, error) {
	var balance gnoland.Balance
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return balance, fmt.Errorf("error creating entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return balance, fmt.Errorf("error generating mnemonic: %w", err)
	}

	var keyInfo keys.Info
	if keyInfo, err = kb.CreateAccount(accountName, mnemonic, "", "", 0, 0); err != nil {
		return balance, fmt.Errorf("unable to create account: %w", err)
	}

	address := keyInfo.GetAddress()
	return gnoland.Balance{
		Address: address,
		Amount:  std.Coins{std.NewCoin("ugnot", 10e6)},
	}, nil
}
