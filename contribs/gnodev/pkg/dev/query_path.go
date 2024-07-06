package dev

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type PackageModifier struct {
	Path    string
	Creator crypto.Address
	Deposit std.Coins
}

type PackageMetaMap struct {
	Creator crypto.Address
	Deposit std.Coins

	queries map[string]PackageModifier
}

func ResolvePackageModifierQuery(bk *address.Book, path string) (PackageModifier, error) {
	var query PackageModifier

	upath, err := url.Parse(path)
	if err != nil {
		return query, fmt.Errorf("malformed path/query: %w", err)
	}

	path = filepath.Clean(upath.Path)

	// Check for creator option
	creator := upath.Query().Get("creator")
	if creator != "" {
		address, err := crypto.AddressFromBech32(creator)
		if err != nil {
			var ok bool
			address, ok = bk.GetByName(creator)
			if !ok {
				return query, fmt.Errorf("invalid name or address for creator %q", creator)
			}
		}

		query.Creator = address
	}

	// Check for deposit option
	deposit := upath.Query().Get("deposit")
	if deposit != "" {
		coins, err := std.ParseCoins(deposit)
		if err != nil {
			return query, fmt.Errorf(
				"unable to parse deposit amount %q (should be in the form xxxugnot): %w", deposit, err,
			)
		}

		query.Deposit = coins
	}

	return query, nil
}
