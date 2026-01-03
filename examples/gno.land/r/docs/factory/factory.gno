package factory

import (
	"chain/runtime"
	"strconv"
	"time"

	"gno.land/p/nt/avl"
)

// Token represents a simple token created by the factory
type Token struct {
	Name    string
	Symbol  string
	Owner   address
	Supply  uint64
	Created int64
}

var (
	tokens     avl.Tree // symbol -> *Token
	tokenCount int
)

// CreateToken creates a new token with the given parameters
func CreateToken(_ realm, name, symbol string, initialSupply uint64) string {
	// Validate inputs
	if name == "" {
		panic("name cannot be empty")
	}
	if symbol == "" || len(symbol) > 8 {
		panic("symbol must be 1-8 characters")
	}
	if tokens.Has(symbol) {
		panic("token with this symbol already exists")
	}

	caller := runtime.PreviousRealm().Address()

	token := &Token{
		Name:    name,
		Symbol:  symbol,
		Owner:   caller,
		Supply:  initialSupply,
		Created: time.Now().Unix(),
	}

	tokens.Set(symbol, token)
	tokenCount++

	return "Token " + symbol + " created successfully!"
}

// GetToken returns information about a token
func GetToken(symbol string) *Token {
	val, ok := tokens.Get(symbol)
	if !ok {
		return nil
	}
	return val.(*Token)
}

// GetTokenCount returns the number of tokens created
func GetTokenCount() int {
	return tokenCount
}

func Render(path string) string {
	if path != "" {
		// Render specific token
		token := GetToken(path)
		if token == nil {
			return "# Token Not Found\n\nNo token with symbol `" + path + "` exists."
		}
		return renderToken(token)
	}

	return renderHome()
}

func renderHome() string {
	out := `# Factory Pattern

The factory pattern allows a realm to create and manage multiple instances
of objects. This is useful for token factories, NFT minting, contract deployment,
and user-generated content.

## How It Works

1. Users call a "create" function with parameters
2. The factory validates the inputs
3. A new object is created and stored in an [AVL tree](/p/nt/avl)
4. The factory tracks all created objects

## This Example

This realm is a simple token factory. It demonstrates creating objects with
validated parameters, storing objects in an AVL tree for efficient lookup,
preventing duplicates (unique symbols), and ownership tracking using
[runtime.PreviousRealm()](/r/docs/crossing).

## Live Demo

### Create a New Token

[Create Token](/r/docs/factory$help&func=CreateToken)

### Existing Tokens

**Total tokens created:** ` + strconv.Itoa(tokenCount) + `

`
	if tokenCount == 0 {
		out += "_No tokens created yet._\n"
	} else {
		tokens.Iterate("", "", func(key string, value any) bool {
			token := value.(*Token)
			out += "- [" + token.Symbol + "](/r/docs/factory:" + token.Symbol + ") - " + token.Name + "\n"
			return false
		})
	}

	out += `
---

## Code Pattern

` + "```go" + `
package factory

import "gno.land/p/nt/avl"

type Product struct {
    ID    string
    Owner address
    // ... other fields
}

var products avl.Tree // id -> *Product

func CreateProduct(_ realm, id string) {
    if products.Has(id) {
        panic("product already exists")
    }
    
    product := &Product{
        ID:    id,
        Owner: runtime.PreviousRealm().Address(),
    }
    
    products.Set(id, product)
}

func GetProduct(id string) *Product {
    val, ok := products.Get(id)
    if !ok {
        return nil
    }
    return val.(*Product)
}
` + "```" + `

[View source code](/r/docs/factory$source)

---

See also:
- [Registry Pattern](/r/docs/registry) - Register and discover entities
- [Crossing](/r/docs/crossing) - Understanding realm boundaries
`
	return out
}

func renderToken(token *Token) string {
	return `# Token: ` + token.Symbol + `

## Details

| Property | Value |
|----------|-------|
| Name | ` + token.Name + ` |
| Symbol | ` + token.Symbol + ` |
| Owner | ` + token.Owner.String() + ` |
| Total Supply | ` + strconv.FormatUint(token.Supply, 10) + ` |

---

[Back to Factory](/r/docs/factory)
`
}
