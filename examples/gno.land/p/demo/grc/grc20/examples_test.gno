// NOTE: The Example tests below are currently commented out since Example tests
// are not yet supported in Gno. They are kept here for inspiration and documentation
// purposes only, and will be uncommented once Example testing is implemented.
// See: https://github.com/gnolang/gno/pull/3353
package grc20

/*
import (
	"std"

	"gno.land/p/demo/ufmt"
)

// ExampleInit demonstrates how to initialize a new token.
func ExampleInit() {
	token, ledger := NewToken("ExampleToken", "EXM", 2)
	ufmt.Println(token.GetName())     // Output: ExampleToken
	ufmt.Println(token.GetSymbol())   // Output: EXM
	ufmt.Println(token.GetDecimals()) // Output: 2
	ufmt.Println(ledger.totalSupply)  // Output: 0
}

// ExampleTransfer demonstrates how to transfer tokens between accounts.
func ExampleTransfer() {
	token, ledger := NewToken("ExampleToken", "EXM", 2)
	tony := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")
	stefan := std.Address("g1dxr7hh6rm3r5wlrkg23qm0fugak3lu8x30kky8")

	ledger.Mint(tony, 1000)
	ledger.Transfer(tony, stefan, 200)

	ufmt.Println(token.BalanceOf(tony))   // Output: 800
	ufmt.Println(token.BalanceOf(stefan)) // Output: 200
}

// ExampleApprove demonstrates how to approve a spender and transfer tokens using allowance.
func ExampleApprove() {
	token, ledger := NewToken("ExampleToken", "EXM", 2)
	tony := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")
	stefan := std.Address("g1dxr7hh6rm3r5wlrkg23qm0fugak3lu8x30kky8")
	carl := std.Address("carl")

	ledger.Mint(tony, 1000)
	ledger.Approve(tony, stefan, 300)
	ledger.TransferFrom(tony, stefan, carl, 200)

	ufmt.Println(token.BalanceOf(tony))         // Output: 800
	ufmt.Println(token.BalanceOf(carl))         // Output: 200
	ufmt.Println(token.Allowance(tony, stefan)) // Output: 100
}

// ExampleMint demonstrates how to mint new tokens.
func ExampleMint() {
	token, ledger := NewToken("ExampleToken", "EXM", 2)
	tony := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")

	ledger.Mint(tony, 500)

	ufmt.Println(token.BalanceOf(tony)) // Output: 500
	ufmt.Println(token.TotalSupply())   // Output: 500
}

// ExampleBurn demonstrates how to burn tokens.
func ExampleBurn() {
	token, ledger := NewToken("ExampleToken", "EXM", 2)
	tony := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")

	ledger.Mint(tony, 500)
	ledger.Burn(tony, 200)

	ufmt.Println(token.BalanceOf(tony)) // Output: 300
	ufmt.Println(token.TotalSupply())   // Output: 300
}

// ExampleCallerTeller demonstrates how to use the CallerTeller.
func ExampleCallerTeller() {
	token, _ := NewToken("ExampleToken", "EXM", 2)
	teller := token.CallerTeller()

	tony := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")
	stefan := std.Address("g1dxr7hh6rm3r5wlrkg23qm0fugak3lu8x30kky8")

	std.TestSetOrigCaller(tony)
	teller.Approve(stefan, 100)

	ufmt.Println(token.Allowance(tony, stefan)) // Output: 100
}

// ExampleReadonlyTeller demonstrates how to use the ReadonlyTeller.
func ExampleReadonlyTeller() {
	token, _ := NewToken("ExampleToken", "EXM", 2)
	teller := token.ReadonlyTeller()

	tony := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")

	// Attempting to approve should result in an error
	err := teller.Approve(tony, 100)
	ufmt.Println(err == ErrReadonly) // Output: true
}

// ExampleRealmTeller demonstrates how to use the RealmTeller.
func ExampleRealmTeller() {
	token, _ := NewToken("ExampleToken", "EXM", 2)
	teller := token.RealmTeller()

	tony := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")
	stefan := std.Address("g1dxr7hh6rm3r5wlrkg23qm0fugak3lu8x30kky8")

	std.TestSetOrigCaller(tony)
	teller.Transfer(stefan, 50)

	ufmt.Println(token.BalanceOf(stefan)) // Output: 50
}

// ExampleImpersonateTeller demonstrates how to use the ImpersonateTeller.
func ExampleImpersonateTeller() {
	token, ledger := NewToken("ExampleToken", "EXM", 2)
	tony := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")
	stefan := std.Address("g1dxr7hh6rm3r5wlrkg23qm0fugak3lu8x30kky8")

	ledger.Mint(tony, 1000)
	teller := ledger.ImpersonateTeller(stefan)

	teller.Transfer(tony, 100)

	ufmt.Println(token.BalanceOf(tony))   // Output: 1100
	ufmt.Println(token.BalanceOf(stefan)) // Output: 0
}

*/

func ExampleInit()                            {}
func ExampleExposeBankForMaketxRunOrImports() {}
func ExampleCustomTellerImpl()                {}
func ExampleAllowance()                       {}
func ExampleRealmBanker()                     {}
func ExamplePrevRealmBanker()                 {}
func ExampleAccountBanker()                   {}
func ExampleTransfer()                        {}
func ExampleApprove()                         {}
func ExampleTransferFrom()                    {}
func ExampleMint()                            {}
func ExampleBurn()                            {}
