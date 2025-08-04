package volos_mocks

import (
	"std"
	"gno.land/r/volos"
	"gno.land/r/demo/grc20reg"
)

// FlashBorrowerMock is a mock contract that demonstrates how to use Volos flash loans
// It implements the FlashLoanCallback interface
type FlashBorrowerMock struct {
	// No need to store volos address since we're in the same package
}

var borrower *FlashBorrowerMock

func init() {
	borrower = &FlashBorrowerMock{}
}

// FlashLoan initiates a flash loan from Volos
func FlashLoan(cur realm, token string, assets int64, data []byte) {
	// Call the Volos flash loan function with our borrower
	volos.FlashLoan(cross, token, assets, data, borrower)
}

// OnVolosFlashLoan is the callback function that gets called during flash loan execution
// This implements the FlashLoanCallback interface
func (f *FlashBorrowerMock) OnVolosFlashLoan(assets int64, data []byte) {
	token := string(data)

	// Approve Volos to take back the tokens using the token's teller
	tokenObj := grc20reg.MustGet(token)
	teller := tokenObj.RealmTeller()
	volosAddr := std.CurrentRealm().Address()
	err := teller.Approve(volosAddr, assets)
	if err != nil {
		panic("approval failed")
	}
}