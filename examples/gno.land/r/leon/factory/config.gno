package factory

import "gno.land/p/demo/ownable"

var (
	MyNFT   *NFT
	MyToken *Token
)

func SetNFT(nft *NFT) {
	MyNFT = nft
}

func SetToken(t *Token) {
	MyToken = t
}

type NFT struct {
	name string
	ownable.Ownable
}

type Token struct {
	name  string
	Owner *ownable.Ownable
}

func NewToken(name string) *Token {
	return &Token{
		name:  name,
		Owner: ownable.NewWithCurrent(),
	}
}

func NewTokenCrossing(name string) *Token {
	crossing()

	return &Token{
		name:  name,
		Owner: ownable.NewWithCurrent(),
	}
}

func (t *Token) ChangeName(newName string) {
	t.Owner.AssertOwnedByCurrent()
	t.name = newName
}

func (t *NFT) ChangeName(newName string) {
	t.AssertOwnedByCurrent()
	t.name = newName
}
