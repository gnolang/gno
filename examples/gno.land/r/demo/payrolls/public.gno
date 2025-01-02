package payrolls

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

const denom = "ugnot"

var (
	id                    seqid.ID
	payrolls              avl.Tree                          // seqid.ID -> *payroll
	payrollsByVault       = make(map[string]*avl.Tree)      // seqid.ID -> *payroll
	payrollsByBeneficiary = make(map[std.Address]*avl.Tree) // seqid.ID -> *payroll
)

func Create(namespace string, label string, beneficiary std.Address, cap int64, distrib DistribFn, breakup BreakupFn) seqid.ID {
	vaultID := vaultIDFromRealm(std.PrevRealm(), namespace)

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
		cap:         cap,
		distrib:     distrib,
		breakup:     breakup,
		beneficiary: beneficiary,
		label:       label,
		creatorAddr: std.PrevRealm().Addr(),
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

func Claim(id seqid.ID, amount int64, destination std.Address) {
	if amount < -1 {
		panic(errors.New("invalid amount"))
	}
	anyAmount := amount == -1
	if amount == 0 {
		panic(errors.New("invalid amount"))
	}

	p := mustGetPayroll(id)
	if std.PrevRealm().Addr() != p.beneficiary {
		panic(errors.New("only beneficiary can withdraw"))
	}

	vaultAmount := getVaultAmount(p.vaultID)
	if vaultAmount == 0 {
		panic(errors.New("vault empty"))
	}
	if anyAmount {
		amount = int64(vaultAmount)
	}
	if !anyAmount && amount > int64(vaultAmount) {
		panic("not enough in vault")
	}

	withdrawn := p.claim(amount)
	if withdrawn == 0 {
		panic(errors.New("nothing to claim"))
	}

	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(std.CurrentRealm().Addr(), destination, std.NewCoins(std.NewCoin(denom, int64(withdrawn))))

	vaults.Set(p.vaultID, vaultAmount-withdrawn)
}

func ClaimAll(destination std.Address) {
	beneficiary := std.PrevRealm().Addr()
	claimed := false
	payrollsByBeneficiary[beneficiary].Iterate("", "", func(key string, value interface{}) bool {
		p := value.(*payroll)
		Claim(p.id, -1, destination)
		claimed = true
		return false
	})
	if !claimed {
		panic(errors.New("nothing to claim"))
	}
}

func Fund(base string, namespace string) {
	std.AssertOriginCall()

	var (
		id  string
		err error
	)

	if base == "" {
		id = vaultIDFromRealm(std.PrevRealm(), namespace)
	} else {
		id, err = vaultID(base, namespace)
		if err != nil {
			panic(err)
		}
	}

	send := std.GetOrigSend()
	amount := send.AmountOf(denom)
	if amount <= 0 {
		panic(errors.New("invalid amount"))
	}
	for _, coin := range send {
		if coin.Denom != denom {
			panic(ufmt.Errorf("only %q supported, got %q", denom, coin.Denom))
		}
	}

	vaults.Set(id, getVaultAmount(id)+uint64(amount))
}

func WithdrawFunds(namespace string, amount int64, destination std.Address) {
	if amount < -1 {
		panic(errors.New("invalid amount"))
	}
	anyAmount := amount == -1
	if amount == 0 {
		panic(errors.New("invalid amount"))
	}

	prevRealm := std.PrevRealm()
	if len(destination) == 0 {
		destination = prevRealm.Addr()
	}

	id := vaultIDFromRealm(prevRealm, namespace)

	vaultAmount := getVaultAmount(id)
	if vaultAmount == 0 {
		panic(errors.New("vault empty"))
	}

	available := int64(vaultAmount)
	payrollsByVault[id].Iterate("", "", func(_ string, value interface{}) bool {
		p := value.(*payroll)
		available -= int64(p.available())
		return false
	})

	if anyAmount {
		amount = available
	} else if available < amount {
		panic(errors.New("not enough in vault"))
	}

	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(std.CurrentRealm().Addr(), destination, std.NewCoins(std.Coin{Denom: denom, Amount: amount}))

	vaults.Set(id, vaultAmount-uint64(amount))
}

func Stop(id seqid.ID) {
	p := mustGetPayroll(id)
	p.stop(std.PrevRealm().Addr())
}

func Pause(id seqid.ID) {
	p := mustGetPayroll(id)
	p.pause(std.PrevRealm().Addr())
}

func Resume(id seqid.ID) {
	p := mustGetPayroll(id)
	p.resume(std.PrevRealm().Addr())
}

func mustGetPayroll(id seqid.ID) *payroll {
	ip, ok := payrolls.Get(id.String())
	if !ok {
		panic(errors.New("no such payroll"))
	}
	return ip.(*payroll)
}
