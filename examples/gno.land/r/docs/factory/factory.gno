package factory

import (
	"chain/runtime/unsafe"
	"errors"
	"strconv"
)

// TokenType represents different types of tokens
type TokenType int

const (
	Fungible TokenType = iota
	NonFungible
	SoulBound
)

// Token is the common interface for all token types
type Token interface {
	GetInfo() string
	GetType() string
	Transfer(to address, amount uint64) error
}

// FungibleToken implements divisible tokens (like coins)
type FungibleToken struct {
	name   string
	amount uint64
	owner  address
}

func (f *FungibleToken) GetInfo() string {
	return f.name + ": " + strconv.FormatUint(f.amount, 10) + " units"
}

func (f *FungibleToken) GetType() string {
	return "Fungible"
}

func (f *FungibleToken) Transfer(to address, amount uint64) error {
	if amount > f.amount {
		return errors.New("insufficient balance")
	}
	f.amount -= amount
	return nil
}

// NonFungibleToken implements unique tokens (like NFTs)
type NonFungibleToken struct {
	name  string
	owner address
}

func (n *NonFungibleToken) GetInfo() string {
	return "NFT: " + n.name
}

func (n *NonFungibleToken) GetType() string {
	return "Non-Fungible"
}

func (n *NonFungibleToken) Transfer(to address, amount uint64) error {
	if amount != 1 {
		return errors.New("NFTs can only be transferred as a whole (amount must be 1)")
	}
	n.owner = to
	return nil
}

// SoulBoundToken implements non-transferable tokens
type SoulBoundToken struct {
	name  string
	owner address
}

func (s *SoulBoundToken) GetInfo() string {
	return "SoulBound: " + s.name
}

func (s *SoulBoundToken) GetType() string {
	return "SoulBound"
}

func (s *SoulBoundToken) Transfer(to address, amount uint64) error {
	return errors.New("soulbound tokens cannot be transferred")
}

// CreateToken is the factory function that creates different token types
func CreateToken(tokenType TokenType, name string, owner address) Token {
	switch tokenType {
	case Fungible:
		return &FungibleToken{name: name, amount: 1000, owner: owner}
	case NonFungible:
		return &NonFungibleToken{name: name, owner: owner}
	case SoulBound:
		return &SoulBoundToken{name: name, owner: owner}
	default:
		panic("unknown token type")
	}
}

// Demo variables
var (
	lastToken Token
)

// CreateFungibleToken creates a fungible token
func CreateFungibleToken(_ realm, name string) string {
	caller := unsafe.PreviousRealm().Address()
	token := CreateToken(Fungible, name, caller)
	lastToken = token
	return "Created: " + token.GetInfo()
}

// CreateNFT creates a non-fungible token
func CreateNFT(_ realm, name string) string {
	caller := unsafe.PreviousRealm().Address()
	token := CreateToken(NonFungible, name, caller)
	lastToken = token
	return "Created: " + token.GetInfo()
}

// CreateSoulBound creates a soul-bound token
func CreateSoulBound(_ realm, name string) string {
	caller := unsafe.PreviousRealm().Address()
	token := CreateToken(SoulBound, name, caller)
	lastToken = token
	return "Created: " + token.GetInfo()
}

func Render(path string) string {
	out := `# Factory Pattern

Creates different types of objects that share a common interface.

## Code Pattern

` + "```go" + `
// Common interface
type Token interface {
    GetInfo() string
    Transfer(to address, amount uint64) error
}

// Concrete types
type FungibleToken struct { name string, amount uint64 }
type NonFungibleToken struct { name string }
type SoulBoundToken struct { name string }

// Factory function
func CreateToken(tokenType int, name string) Token {
    switch tokenType {
    case Fungible:
        return &FungibleToken{name: name, amount: 1000}
    case NonFungible:
        return &NonFungibleToken{name: name}
    case SoulBound:
        return &SoulBoundToken{name: name}
    }
}

// Work with interface, not concrete types
token := CreateToken(Fungible, "USDC")
if err := token.Transfer(recipientAddr, 100); err != nil {
    panic(err)
}
` + "```" + `

## Live Demo

- [Create Fungible Token](/r/docs/factory$help&func=CreateFungibleToken)
- [Create NFT](/r/docs/factory$help&func=CreateNFT)
- [Create SoulBound Token](/r/docs/factory$help&func=CreateSoulBound)

`

	if lastToken != nil {
		out += "### Last Created Token\n\n"
		out += "**Type:** " + lastToken.GetType() + "\n\n"
		out += "**Info:** " + lastToken.GetInfo() + "\n\n"
	}

	out += `
`
	return out
}
