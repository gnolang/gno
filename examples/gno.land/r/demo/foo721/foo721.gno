package nft

import (
	"bytes"
	"std"
	"strconv"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
)

var token *grc721.Token
var o *ownable.Ownable

var idCounter int
var ids []int

func init() {
	token = grc721.NewGRC721Token("Example NFT", "EX")
	o = ownable.New()

	Mint(o.Owner(), strconv.Itoa(idCounter))
}

func BalanceOf(owner std.Address) uint64 {
	return token.BalanceOf(owner)
}

func OwnerOf(tokenID string) std.Address {
	return token.OwnerOf(tokenID)
}

func TransferFrom(from, to std.Address, tokenID string) {
	token.TransferFrom(from, to, tokenID)
}

func Approve(to std.Address, tokenID string) {
	token.Approve(to, tokenID)
}

func SetApprovalForAll(operator std.Address, approved bool) {
	token.SetApprovalForAll(operator, approved)
}

func GetApproved(tokenID string) std.Address {
	return token.GetApproved(tokenID)
}

func IsApprovedForAll(owner, operator std.Address) bool {
	return token.IsApprovedForAll(owner, operator)
}

func TokenURI(tokenID string) string {
	return token.TokenURI(tokenID)
}

func SetTokenURI(tokenID string, uri string) string {
	if err := o.CallerIsOwner(); err != nil {
		panic("only owner can mint NFTs")
	}

	return token.SetTokenURI(tokenID, uri)
}

func Name() string {
	return token.Name()
}

func Symbol() string {
	return token.Symbol()
}

func Mint(to std.Address, tokenID string) {
	if err := o.CallerIsOwner(); err != nil {
		panic("only owner can mint NFTs")
	}

	token.Mint(to, tokenID)

	ids = append(ids, idCounter)
	idCounter += 1
}

func GetCollectionOwner() std.Address {
	return o.Owner()
}

// Render renders tokens & their owners
func Render(_ string) string {
	var buf bytes.Buffer

	buf.WriteString(ufmt.Sprintf("# NFT Collection: \"%s\" $%s\n\n", Name(), Symbol()))

	for i := 0; i < len(ids); i++ {
		owner := token.OwnerOf(strconv.Itoa(ids[i]))
		str := ufmt.Sprintf("#### TokenID #%d - owned by %s\n", ids[i], owner.String())
		buf.WriteString(str)
	}

	return buf.String()
}
