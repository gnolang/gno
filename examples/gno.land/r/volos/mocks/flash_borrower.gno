package mocks

import (
	"std"
	volos "gno.land/r/volos/core"
	"gno.land/r/demo/grc20reg"
)

// FlashBorrowerMock is a mock contract that demonstrates how to use Volos flash loans
// It implements the FlashLoanCallback interface
type FlashBorrowerMock struct {
	volosAddr std.Address
}

var borrower *FlashBorrowerMock

func init() {
	borrower = &FlashBorrowerMock{
		volosAddr: std.DerivePkgAddr("gno.land/r/volos"),
	}
}

// FlashLoan initiates a flash loan from Volos
func FlashLoan(cur realm, token string, assets int64, data []byte) {
	// Call the Volos flash loan function with our borrower
	volos.FlashLoan(cross, token, assets, data, borrower)
}

// OnVolosFlashLoan is the callback function that gets called during flash loan execution
// This implements the FlashLoanCallback interface
func (f *FlashBorrowerMock) OnVolosFlashLoan(cur realm, assets int64, data []byte) {
	token := string(data)

	// Approve Volos to take back the tokens using the token's teller
	tokenObj := grc20reg.MustGet(token)
	teller := tokenObj.RealmTeller()
	err := teller.Approve(f.volosAddr, assets)
	if err != nil {
		panic("approval failed")
	}
}