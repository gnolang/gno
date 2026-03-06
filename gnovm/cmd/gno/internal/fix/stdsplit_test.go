package fix

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdsplit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "simple std import removal",
			input: `package test
import "std"
func main() {}`,
			expected: `package test

func main() {}`,
		},
		{
			name: "std.Coin rewrite",
			input: `package test
import "std"
func main() {
	addr := std.Coin{}
}`,
			expected: `package test

import "chain"

func main() {
	addr := chain.Coin{}
}`,
		},
		{
			name: "std.AssertOriginCall rewrite",
			input: `package test
import "std"
func main() {
	std.AssertOriginCall()
}`,
			expected: `package test

import "chain/runtime"

func main() {
	runtime.AssertOriginCall()
}`,
		},
		{
			name: "multiple std functions",
			input: `package test
import "std"
func main() {
	addr := std.Address("123")
	_ = std.Coin{}
	std.AssertOriginCall()
}`,
			expected: `package test

import (
	"chain"
	"chain/runtime"
)

func main() {
	addr := address("123")
	_ = chain.Coin{}
	runtime.AssertOriginCall()
}`,
		},
		{
			name: "existing imports",
			input: `package test
import (
	"fmt"
	"std"
)
func main() {
	fmt.Println("hello")
	c := std.Coin{}
}`,
			expected: `package test

import (
	"chain"
	"fmt"
)

func main() {
	fmt.Println("hello")
	c := chain.Coin{}
}`,
		},
		{
			name: "aliased import",
			input: `package test
import s "std"
func main() {
	c := s.Coin{}
}`,
			expected: `package test

import "chain"

func main() {
	c := chain.Coin{}
}`,
		},
		{
			name: "handle RawAddressSize special case",
			input: `package test
import "std"
func main() {
	size := std.RawAddressSize
}`,
			expected: `package test

func main() {
	size := 20
}`,
		},
		{
			name: "name collision handling",
			input: `package test
import "std"
func main() {
	// Local variable with same name as import identifier
	runtime := "collision"
	// Should still rewrite properly
	std.AssertOriginCall()
	println(runtime)
}`,
			expected: `package test

import "chain/runtime"

func main() {
	// Local variable with same name as import identifier
	runtime_ := "collision"
	// Should still rewrite properly
	runtime.AssertOriginCall()
	println(runtime_)
}`,
		},
		{
			name: "multiple std functions with same target package",
			input: `package test
import "std"
func main() {
	std.AssertOriginCall()
	std.PreviousRealm()
	std.CurrentRealm()
}`,
			expected: `package test

import "chain/runtime"

func main() {
	runtime.AssertOriginCall()
	runtime.PreviousRealm()
	runtime.CurrentRealm()
}`,
		},
		{
			name: "std with aliased target package",
			input: `package test
import (
	"std"
	rt "chain/runtime"
)
func main() {
	std.AssertOriginCall()
	rt.CurrentRealm()
}`,
			expected: `package test

import rt "chain/runtime"

func main() {
	rt.AssertOriginCall()
	rt.CurrentRealm()
}`,
		},
		{
			name: "fix with nested blocks and scopes",
			input: `package test
import "std"
func main() {
	if true {
		addr := std.Address("g1337")
		if true {
			std.AssertOriginCall()
		}
	}
	for i := 0; i < 10; i++ {
		std.Emit("test", i)
	}
}`,
			expected: `package test

import (
	"chain"
	"chain/runtime"
)

func main() {
	if true {
		addr := address("g1337")
		if true {
			runtime.AssertOriginCall()
		}
	}
	for i := 0; i < 10; i++ {
		chain.Emit("test", i)
	}
}`,
		},
		{
			name: "handle aliased functions (GetOrigCaller to OriginCaller)",
			input: `package test
import "std"
func main() {
	caller := std.GetOrigCaller()
}`,
			expected: `package test

import "chain/runtime"

func main() {
	caller := runtime.OriginCaller()
}`,
		},
		{
			name: "handle banker functions",
			input: `package test
import "std"
func main() {
	banker := std.NewBanker(std.BankerTypeReadonly)
	send := std.OriginSend()
}`,
			expected: `package test

import "chain/banker"

func main() {
	banker_ := banker.NewBanker(banker.BankerTypeReadonly)
	send := banker.OriginSend()
}`,
		},
		{
			name: "scope shadowing name collision",
			input: `package test
import "std"
func main() {
	banker := "someone who banks"
	{
		// In this scope, banker is shadowed but we need to rename
		// the variable to avoid collision
		std.NewBanker(std.BankerTypeReadonly)
	}
	println(banker)
}`,
			expected: `package test

import "chain/banker"

func main() {
	banker_ := "someone who banks"
	{
		// In this scope, banker is shadowed but we need to rename
		// the variable to avoid collision
		banker.NewBanker(banker.BankerTypeReadonly)
	}
	println(banker_)
}`,
		},
		{
			name: "allow methods of same name",
			input: `package main

import "std"

type (
	S string
	I int
)

func (S) String() { return string(std.DerivePkgAddr("123")) }
func (I) String() { return std.RawAddressSize + 123 }`,
			expected: `package main

import "chain"

type (
	S string
	I int
)

func (S) String() { return string(chain.PackageAddress("123")) }
func (I) String() { return 20 + 123 }`,
		},
		{
			name: "colliding imports",
			input: `package main

import (
	"std"
)

var (
	banker  = 123
	banker_ = 456
)

func main() {
	println(std.OriginSend())
}`,
			expected: `package main

import banker__ "chain/banker"

var (
	banker  = 123
	banker_ = 456
)

func main() {
	println(banker__.OriginSend())
}`,
		},
		{
			name: "colliding imports in func scope",
			input: `package main

import (
	"std"
)

func main() {
	banker := 123
	banker_ := 456
	_ = 123 + banker + banker_
	println(std.OriginSend())
}`,
			expected: `package main

import "chain/banker"

func main() {
	banker__ := 123
	banker_ := 456
	_ = 123 + banker__ + banker_
	println(banker.OriginSend())
}`,
		},
		{
			name: "shadowing after first use",
			input: `package disperse

import (
	"std"

	tokens "gno.land/r/demo/grc20factory"
)

// Get address of Disperse realm
var realmAddr = std.CurrentRealm().Address()

// DisperseUgnot parses receivers and amounts and sends out ugnot
// The function will send out the coins to the addresses and return the leftover coins to the caller
// if there are any to return
func DisperseUgnot(addresses []std.Address, coins std.Coins) {
	coinSent := std.OriginSend()
	caller := std.PreviousRealm().Address()
	banker := std.NewBanker(std.BankerTypeOriginSend)

	if len(addresses) != len(coins) {
		panic(ErrNumAddrValMismatch)
	}

	for _, coin := range coins {
		if coin.Amount <= 0 {
			panic(ErrNegativeCoinAmount)
		}

		if banker.GetCoins(realmAddr).AmountOf(coin.Denom) < coin.Amount {
			panic(ErrMismatchBetweenSentAndParams)
		}
	}

	// Send coins
	for i, _ := range addresses {
		banker.SendCoins(realmAddr, addresses[i], std.NewCoins(coins[i]))
	}

	// Return possible leftover coins
	for _, coin := range coinSent {
		leftoverAmt := banker.GetCoins(realmAddr).AmountOf(coin.Denom)
		if leftoverAmt > 0 {
			send := chain.Coins{chain.NewCoin(coin.Denom, leftoverAmt)}
			banker.SendCoins(realmAddr, caller, send)
		}
	}
}`,
			expected: `package disperse

import (
	"chain"
	"chain/banker"
	"chain/runtime"

	tokens "gno.land/r/demo/grc20factory"
)

// Get address of Disperse realm
var realmAddr = runtime.CurrentRealm().Address()

// DisperseUgnot parses receivers and amounts and sends out ugnot
// The function will send out the coins to the addresses and return the leftover coins to the caller
// if there are any to return
func DisperseUgnot(addresses []address, coins chain.Coins) {
	coinSent := banker.OriginSend()
	caller := runtime.PreviousRealm().Address()
	banker_ := banker.NewBanker(banker.BankerTypeOriginSend)

	if len(addresses) != len(coins) {
		panic(ErrNumAddrValMismatch)
	}

	for _, coin := range coins {
		if coin.Amount <= 0 {
			panic(ErrNegativeCoinAmount)
		}

		if banker_.GetCoins(realmAddr).AmountOf(coin.Denom) < coin.Amount {
			panic(ErrMismatchBetweenSentAndParams)
		}
	}

	// Send coins
	for i, _ := range addresses {
		banker_.SendCoins(realmAddr, addresses[i], chain.NewCoins(coins[i]))
	}

	// Return possible leftover coins
	for _, coin := range coinSent {
		leftoverAmt := banker_.GetCoins(realmAddr).AmountOf(coin.Denom)
		if leftoverAmt > 0 {
			send := chain.Coins{chain.NewCoin(coin.Denom, leftoverAmt)}
			banker_.SendCoins(realmAddr, caller, send)
		}
	}
}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tc.input, parser.ParseComments)
			require.NoError(t, err)

			stdsplit(f)

			// Convert the AST back to source code for comparison
			output := astToString(t, fset, f)
			assert.Equal(t, tc.expected, output)
		})
	}
}

func astToString(t *testing.T, fset *token.FileSet, f *ast.File) string {
	t.Helper()
	var buf bytes.Buffer
	err := format.Node(&buf, fset, f)
	require.NoError(t, err)
	return strings.TrimSuffix(buf.String(), "\n")
}

func TestStdsplitFixIndicator(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedFixed bool
	}{
		{
			name: "file without std import",
			input: `package test
import "fmt"
func main() {
	fmt.Println("hello")
}`,
			expectedFixed: false,
		},
		{
			name: "file with std import but no std calls",
			input: `package test
import (
	"fmt"
	"std"
)
func main() {
	fmt.Println("hello")
}`,
			expectedFixed: true, // Because the std import is removed
		},
		{
			name: "file with std calls",
			input: `package test
import "std"
func main() {
	std.Address("g1337")
}`,
			expectedFixed: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tc.input, parser.ParseComments)
			require.NoError(t, err)

			fixed := stdsplit(f)
			assert.Equal(t, tc.expectedFixed, fixed)
		})
	}
}

func TestStdsplitWithMixedPackages(t *testing.T) {
	input := `package test
import (
	"std"
	"chain"
	"chain/runtime"
)
func main() {
	// Already using chain packages along with std
	std.DerivePkgAddr("g1337")
	chain.Emit("event", "data")
	runtime.CurrentRealm()

	// These should be converted
	std.AssertOriginCall()
	std.DerivePkgAddr("g1337")
}`

	expected := `package test

import (
	"chain"
	"chain/runtime"
)

func main() {
	// Already using chain packages along with std
	chain.PackageAddress("g1337")
	chain.Emit("event", "data")
	runtime.CurrentRealm()

	// These should be converted
	runtime.AssertOriginCall()
	chain.PackageAddress("g1337")
}`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", input, parser.ParseComments)
	require.NoError(t, err)

	fixed := stdsplit(f)
	assert.True(t, fixed)

	output := astToString(t, fset, f)
	assert.Equal(t, expected, output)
}
