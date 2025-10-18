// package payrolls is a payroll management example
package payrolls

import (
	"chain"
	"chain/banker"
	"chain/runtime"
	"errors"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/seqid"
	"gno.land/r/demo/grc20reg"
)

var (
	id                    seqid.ID
	payrolls              avl.Tree                      // seqid.ID -> *payroll
	payrollsByVault       = make(map[string]*avl.Tree)  // vaultID -> seqid.ID -> *payroll
	payrollsByBeneficiary = make(map[address]*avl.Tree) // addr -> seqid.ID -> *payroll
)

// Create creates a new payroll using the vault identified by the caller and namespace.
//
// - distrib is the function used to compute the amount of coins to release at a specific point it time. It must not be nil. Example: [MonthlyContinuous]
//
// - breakup is the function used to manage bonuses when a payroll is terminated. It can be nil. Example: [BreakupCDI]
func Create(namespace string, label string, beneficiary address, distrib DistribFn, breakup BreakupFn) seqid.ID {
	assertValidNamespace(namespace)

	vaultID := vaultIDFromRealm(runtime.PreviousRealm(), namespace)

	if distrib == nil {
		panic(errors.New("nil distrib"))
	}
	if !beneficiary.IsValid() {
		panic(errors.New("invalid beneficiary"))
	}

	p := payroll{
		id:          id.Next(),
		vaultID:     vaultID,
		createdAt:   time.Now(),
		distrib:     distrib,
		breakup:     breakup,
		beneficiary: beneficiary,
		label:       label,
		creatorAddr: runtime.PreviousRealm().Addr(),
	}

	payrolls.Set(id.String(), &p)

	if _, ok := payrollsByBeneficiary[p.beneficiary]; !ok {
		payrollsByBeneficiary[p.beneficiary] = avl.NewTree()
	}
	payrollsByBeneficiary[p.beneficiary].Set(p.id.String(), &p)

	if _, ok := payrollsByVault[p.vaultID]; !ok {
		payrollsByVault[p.vaultID] = avl.NewTree()
	}
	payrollsByVault[p.vaultID].Set(p.id.String(), &p)

	return p.id
}

// ClaimAll claims the coins released in the payroll identified by id
//
// - The caller must be the payroll's beneficiary
//
// - If destination is empty, it will send coins to the caller address
func Claim(id seqid.ID, destination address) {
	claimInternal(id, destination, true)
}

func claimInternal(id seqid.ID, destination address, mustClaim bool) bool {
	p := mustGetPayroll(id)
	caller := runtime.PreviousRealm().Addr()
	if caller != p.beneficiary {
		panic(errors.New("only beneficiary can withdraw"))
	}

	vaultAmount := getVaultCoins(p.vaultID)
	claimed := p.claim(vaultAmount)
	if coinsEmpty(claimed) {
		if mustClaim {
			panic(errors.New("nothing to claim"))
		} else {
			return false
		}
	}

	if len(destination) == 0 {
		destination = caller
	}

	sendCoins(destination, claimed)

	vaults.Set(p.vaultID, subCoins(vaultAmount, claimed))

	return true
}

// ClaimAll claims the coins released in all payrolls where the caller is the beneficiary
//
// - If destination is empty, it will send coins to the caller address
func ClaimAll(destination address) {
	beneficiary := runtime.PreviousRealm().Addr()
	claimed := false
	payrollsByBeneficiary[beneficiary].Iterate("", "", func(key string, value interface{}) bool {
		p := value.(*payroll)
		if ok := claimInternal(p.id, destination, false); ok {
			claimed = true
		}
		return false
	})
	if !claimed {
		panic(errors.New("nothing to claim"))
	}
}

// FundNative funds a vault identified by base and namespace using the OrigSend coins
//
// - base must be an address or a pkgpath
//
// - Anyone can fund any vault
func FundNative(base string, namespace string) {
	assertValidNamespace(namespace)

	runtime.AssertOriginCall()

	var (
		id  string
		err error
	)

	if base == "" {
		id = vaultIDFromRealm(runtime.PreviousRealm(), namespace)
	} else {
		id, err = vaultID(base, namespace)
		if err != nil {
			panic(err)
		}
	}

	send := banker.OriginSend()
	if coinsEmpty(send) {
		panic(errors.New("invalid amount"))
	}
	for i := range send {
		send[i].Denom = "/native/" + send[i].Denom
	}
	vaults.Set(id, addCoins(getVaultCoins(id), send))
}

// FundGRC20Reg funds a vault identified by base and namespace using the fqname of a registered coin
//
// - If amount is empty, it will use the full allowance
//
// - base must be an address or a pkgpath
//
// - Anyone can fund any vault
func FundGRC20Reg(base string, namespace string, fqname string, amount int64) {
	assertValidNamespace(namespace)

	tokenGetter := grc20reg.MustGet(fqname)
	token := tokenGetter()
	if token == nil {
		panic(errors.New("token getter returned nil"))
	}

	fundGRC20(base, namespace, token, "/grc20/"+fqname, amount)
}

func fundGRC20(base string, namespace string, token *grc20.Token, tokenKey string, amount int64) {
	anyAmount := amount == -1
	if !anyAmount && amount <= 0 {
		panic(errors.New("invalid amount"))
	}

	var (
		id  string
		err error
	)

	if base == "" {
		id = vaultIDFromRealm(runtime.PreviousRealm(), namespace)
	} else {
		id, err = vaultID(base, namespace)
		if err != nil {
			panic(err)
		}
	}

	if anyAmount {
		amount = int64(token.Allowance(runtime.PreviousRealm().Addr(), runtime.CurrentRealm().Addr()))
	}

	teller := token.CallerTeller()
	if err := teller.TransferFrom(runtime.PreviousRealm().Addr(), runtime.CurrentRealm().Addr(), uint64(amount)); err != nil {
		panic(err)
	}

	send := chain.NewCoins(chain.NewCoin(tokenKey, int64(amount)))
	vaults.Set(id, addCoins(getVaultCoins(id), send))
}

// WithdrawFunds withdraws from a vault identified by the caller and namespace. It uses the coins identified by a payrolls denom (see [Coins])
//
// - If amount is empty, it will withdraw all available funds
//
// - If destination is empty, it will send to the caller's address
func WithdrawFunds(namespace string, denom string, amount int64, destination address) {
	assertValidNamespace(namespace)

	if amount < -1 {
		panic(errors.New("invalid amount"))
	}
	anyAmount := amount == -1
	if amount == 0 {
		panic(errors.New("invalid amount"))
	}

	prevRealm := runtime.PreviousRealm()
	if len(destination) == 0 {
		destination = prevRealm.Addr()
	}

	id := vaultIDFromRealm(prevRealm, namespace)

	vaultAmount := getVaultCoins(id)
	if vaultAmount.AmountOf(denom) == 0 {
		panic(errors.New("vault empty"))
	}

	available := vaultAmount
	payrollsByVault[id].Iterate("", "", func(_ string, value interface{}) bool {
		p := value.(*payroll)
		available = subCoins(available, p.available())
		return false
	})

	availableAmount := available.AmountOf(denom)
	if anyAmount {
		amount = availableAmount
	} else if availableAmount < amount {
		panic(errors.New("not enough in vault"))
	}

	toSend := chain.NewCoins(chain.Coin{Denom: denom, Amount: amount})
	sendCoins(destination, toSend)

	vaults.Set(id, subCoins(vaultAmount, toSend))
}

// Stop terminates the payroll identified by id. It will distribute coins using the payroll's breakup function.
func Stop(id seqid.ID) {
	p := mustGetPayroll(id)
	p.stop(runtime.PreviousRealm().Addr())
}

// Pause pauses the payroll identified by id. The payroll will not release coins during the pause.
func Pause(id seqid.ID) {
	p := mustGetPayroll(id)
	p.pause(runtime.PreviousRealm().Addr())
}

// Resume resumes the paused payroll identified by id.
func Resume(id seqid.ID) {
	p := mustGetPayroll(id)
	p.resume(runtime.PreviousRealm().Addr())
}

func mustGetPayroll(id seqid.ID) *payroll {
	ip, ok := payrolls.Get(id.String())
	if !ok {
		panic(errors.New("no such payroll"))
	}
	return ip.(*payroll)
}
