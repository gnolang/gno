package payrolls

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

// TODO: support grc20

type DistribFn func(elapsed time.Duration) uint64

const denom = "ugnot"

type payroll struct {
	id             seqid.ID
	vaultID        string
	createdAt      time.Time
	amount         int64
	beneficiary    std.Address
	totalWithdrawn uint64
	distrib        DistribFn
	paused         bool
	pausedAt       time.Time
	pauseDuration  time.Duration
	label          string
}

func Create(namespace string, label string, beneficiary std.Address, amount int64, distrib DistribFn) seqid.ID {
	vaultID := vaultIDFromRealm(std.PrevRealm(), namespace)

	p := payroll{
		id:          id.Next(),
		vaultID:     vaultID,
		createdAt:   time.Now(),
		amount:      amount,
		distrib:     distrib,
		beneficiary: beneficiary,
		label:       label,
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

func CreateMonthly(namespace string, label string, beneficiary std.Address, amount uint64, totalAmount int64) seqid.ID {
	return Create(namespace, label, beneficiary, totalAmount, MonthlyContinuous(amount))
}

func Claim(id seqid.ID, requestedAmount int64, destination std.Address) {
	if requestedAmount == 0 {
		panic(errors.New("invalid amount"))
	}

	p := mustGetPayroll(id)
	if std.PrevRealm().Addr() != p.beneficiary {
		panic(errors.New("only beneficiary can withdraw"))
	}

	vaultAmount := getVaultAmount(p.vaultID)
	if vaultAmount == 0 && requestedAmount != -1 {
		panic(errors.New("nothing to claim: vault empty"))
	}
	if requestedAmount > int64(vaultAmount) {
		requestedAmount = int64(vaultAmount)
	}

	withdrawn := p.withdraw(requestedAmount)
	if withdrawn == 0 && requestedAmount != -1 {
		panic(errors.New("nothing to claim"))
	}

	if withdrawn == 0 {
		return
	}

	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(std.CurrentRealm().Addr(), destination, std.NewCoins(std.NewCoin(denom, int64(withdrawn))))

	vaults.Set(p.vaultID, vaultAmount-withdrawn)
}

func ClaimAll(destination std.Address) {
	// TODO: optimize
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

func Fund(input string, namespace string) {
	std.AssertOriginCall()

	var (
		id  string
		err error
	)

	if input == "" {
		id = vaultIDFromRealm(std.PrevRealm(), namespace)
	} else {
		id, err = vaultID(input, namespace)
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

func Estimate(id seqid.ID, when time.Time) uint64 {
	p := mustGetPayroll(id)
	return p.freeAt(when) - p.totalWithdrawn
}

var id seqid.ID

var payrolls avl.Tree                                       // seqid.ID -> *payroll
var payrollsByVault = make(map[string]*avl.Tree)            // seqid.ID -> *payroll
var payrollsByBeneficiary = make(map[std.Address]*avl.Tree) // seqid.ID -> *payroll

func (p *payroll) freeAt(when time.Time) uint64 {
	var elapsed time.Duration
	if p.paused {
		elapsed = p.pausedAt.Sub(p.createdAt)
	} else {
		elapsed = when.Sub(p.createdAt)
	}
	elapsed -= p.pauseDuration
	res := p.distrib(elapsed)
	if p.amount > 0 && res > uint64(p.amount) {
		return uint64(p.amount)
	}
	return res
}

func (p *payroll) available() uint64 {
	return p.freeAt(time.Now()) - p.totalWithdrawn
}

func (p *payroll) withdraw(amount int64) uint64 {
	available := int64(p.available())
	if amount > available || amount < 0 {
		amount = available
	}
	p.totalWithdrawn += uint64(amount)
	return uint64(amount)
}

func (p *payroll) pause() {
	if p.paused {
		panic(errors.New("already paused"))
	}
	p.paused = true
	p.pausedAt = time.Now()
}

func (p *payroll) resume() {
	if !p.paused {
		panic(errors.New("not paused"))
	}
	p.paused = false
	p.pauseDuration += time.Now().Sub(p.pausedAt)
}

func mustGetPayroll(id seqid.ID) *payroll {
	ip, ok := payrolls.Get(id.String())
	if !ok {
		panic(errors.New("no such payroll"))
	}
	return ip.(*payroll)
}

func StepDistrib(blockAmount uint64, per time.Duration) DistribFn {
	return func(elapsed time.Duration) uint64 {
		blocksCount := elapsed / per
		return blockAmount * uint64(blocksCount)
	}
}

func LinearDistrib(blockAmount uint64, per time.Duration) DistribFn {
	return func(elapsed time.Duration) uint64 {
		return (uint64(elapsed.Seconds()) * blockAmount) / uint64(per.Seconds())
	}
}

func Monthly(amount uint64) DistribFn {
	return StepDistrib(amount, time.Hour*24*30)
}

func MonthlyContinuous(amount uint64) DistribFn {
	return LinearDistrib(amount, time.Hour*24*30)
}

// Usage Create(std.Address("foo"), MonthlyContinuous(1000))
