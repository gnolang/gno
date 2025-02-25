package dev

import (
	"fmt"
	"net/url"
	"path"

	"github.com/gnolang/gno/contribs/gnodev/pkg/address"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type QueryPath struct {
	Path    string
	Creator crypto.Address
	Deposit std.Coins
}

func ResolveQueryPath(bk *address.Book, query string) (QueryPath, error) {
	var qpath QueryPath

	upath, err := url.Parse(query)
	if err != nil {
		return qpath, fmt.Errorf("malformed path/query: %w", err)
	}

	qpath.Path = path.Clean(upath.Path)

	// Check for creator option
	creator := upath.Query().Get("creator")
	if creator != "" {
		address, err := crypto.AddressFromBech32(creator)
		if err != nil {
			var ok bool
			address, ok = bk.GetByName(creator)
			if !ok {
				return qpath, fmt.Errorf("invalid name or address for creator %q", creator)
			}
		}

		qpath.Creator = address
	}

	// Check for deposit option
	deposit := upath.Query().Get("deposit")
	if deposit != "" {
		coins, err := std.ParseCoins(deposit)
		if err != nil {
			return qpath, fmt.Errorf(
				"unable to parse deposit amount %q (should be in the form xxxugnot): %w", deposit, err,
			)
		}

		qpath.Deposit = coins
	}

	return qpath, nil
}
