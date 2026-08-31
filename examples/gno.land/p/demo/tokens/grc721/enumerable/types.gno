// Package enumerable is a stackable GRC721 extension implementing
// EIP-721Enumerable. Attach it via NewEnumerable before the first mint.
package enumerable

import (
	"errors"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/avl/v0"
)

var ErrIndexOutOfRange = errors.New("index out of range")

// Stored by pointer so in-place swap-and-pop updates persist.
type tokenList struct {
	ids []grc721.TokenID
}

type storage struct {
	allTokens  []grc721.TokenID
	allIndex   avl.Tree // grc721.TokenID -> int (position in allTokens)
	owned      avl.Tree // owner address -> *tokenList
	ownedIndex avl.Tree // grc721.TokenID -> int (position within its owner list)
	owner      avl.Tree // grc721.TokenID -> owner address
}

type Enumerable struct {
	core *grc721.Token
	st   *storage
}

type Ledger struct {
	core *grc721.Token
	st   *storage
}

var _ grc721.Extension = (*Ledger)(nil)

var _ grc721.ExtensionView = (*Enumerable)(nil)
