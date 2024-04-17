package main

import (
	"fmt"
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/keyerror"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

func setupKeybase(logger *slog.Logger, cfg *devCfg) (keys.Keybase, error) {
	kb := keys.NewInMemory()

	// Check for home folder
	if cfg.home == "" {
		logger.Warn("local keybase disabled")
	} else if !osm.DirExists(cfg.home) {
		logger.Warn("keybase directory does not exist, no local keys will be imported",
			"path", cfg.home)
	} else if err := importKeybase(kb, cfg.home); err != nil {
		return nil, fmt.Errorf("unable to import local keybase %q: %w", cfg.home, err)
	}

	// Add additional accounts to our keybase
	for acc := range cfg.additionalAccounts {
		// Check if the account exist in the local keybase
		if ok, _ := kb.HasByName(acc); ok {
			continue
		}

		// Check if we have a valid bech32 address instead
		addr, err := crypto.AddressFromBech32(acc)
		if err != nil {
			return nil, fmt.Errorf("invalid bech32 address or unkown key %q", acc)
		}

		// If we already know this address from keybase, skip it
		ok, err := kb.HasByAddress(addr)
		if ok {
			continue
		}

		// We don't know this address, then add it to our keybase
		pub, err := crypto.PubKeyFromBech32(acc)
		if err != nil {
			return nil, fmt.Errorf("unable to get PubKey from %q: %w", acc, err)
		}

		name := fmt.Sprintf("_account#%.6s", addr.String())
		info, err := kb.CreateOffline(name, pub)
		if err != nil {
			return nil, fmt.Errorf("unable to add additional account: %w", err)
		}

		logger.Info("additional account added",
			"name", info.GetName(),
			"addr", info.GetAddress())
	}

	// Ensure that we have a default address
	info, err := kb.GetByAddress(DefaultCreatorAddress)
	switch {
	case err == nil: // Account already exist in the keybase
		logger.Info("default address imported from keybase", "name", info.GetName(), "addr", info.GetAddress())
	case keyerror.IsErrKeyNotFound(err):
		// If the key isn't found, create a default one
		creatorName := fmt.Sprintf("_default#%.6s", DefaultCreatorAddress.String())
		if ok, _ := kb.HasByName(creatorName); ok {
			return nil, fmt.Errorf("unable to create default account, %q already exist in imported keybase", creatorName)
		}

		info, err = kb.CreateAccount(creatorName, DefaultCreatorSeed, "", "", 0, 0)
		if err != nil {
			return nil, fmt.Errorf("unable to create default account %q: %w", DefaultCreatorName, err)
		}

		logger.Warn("default address created",
			"name", info.GetName(),
			"addr", info.GetAddress(),
			"mnemonic", DefaultCreatorSeed,
		)
	default:
		return nil, fmt.Errorf("unable to get address %q: %w", info.GetAddress(), err)
	}

	return kb, nil
}

func importKeybase(to keys.Keybase, path string) error {
	// Load home keybase into our inMemory keybase
	from, err := keys.NewKeyBaseFromDir(path)
	if err != nil {
		return fmt.Errorf("unable to load keybase: %w", err)
	}

	keys, err := from.List()
	if err != nil {
		return fmt.Errorf("unable to list keys: %w", err)
	}

	for _, key := range keys {
		name := key.GetName()
		to.CreateOffline(name, key.GetPubKey())
	}

	return nil
}
