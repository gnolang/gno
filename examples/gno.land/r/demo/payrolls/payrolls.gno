package payrolls

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

type DistribFn func(elapsed time.Duration) uint64
type BreakupFn func(elapsed time.Duration, pauseDuration time.Duration, source BreakupSource) int64

const denom = "ugnot"

type BreakupSource uint

const (
	BreakupSourceCreator BreakupSource = iota
	BreakupSourceBeneficiary
)

type payroll struct {
	id             seqid.ID
	creatorAddr    std.Address
	vaultID        string
	createdAt      time.Time
	cap            int64 // XXX: maybe we should remove this and handle cap in distrib fn directly
	beneficiary    std.Address
	totalWithdrawn uint64
	distrib        DistribFn
	breakup        BreakupFn
	paused         bool
	pausedAt       time.Time
	pauseDuration  time.Duration
	label          string
	breakupAmount  int64
	stopped        bool
	stoppedAt      time.Time
}

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

	withdrawn := p.claim(requestedAmount)
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

var id seqid.ID

var payrolls avl.Tree                                       // seqid.ID -> *payroll
var payrollsByVault = make(map[string]*avl.Tree)            // seqid.ID -> *payroll
var payrollsByBeneficiary = make(map[std.Address]*avl.Tree) // seqid.ID -> *payroll

func (p *payroll) freeAt(when time.Time) uint64 {
	if p.stopped && p.stoppedAt.Before(when) {
		when = p.stoppedAt
	}

	var elapsed time.Duration
	if p.paused {
		elapsed = p.pausedAt.Sub(p.createdAt)
	} else {
		elapsed = when.Sub(p.createdAt)
	}
	elapsed -= p.pauseDuration
	res := p.distrib(elapsed)

	if p.cap > 0 && res > uint64(p.cap) {
		return uint64(p.cap)
	}

	return res
}

func (p *payroll) available() uint64 {
	return (p.freeAt(time.Now()) + uint64(p.breakupAmount)) - p.totalWithdrawn
}

func (p *payroll) estimateDay() uint64 {
	return p.freeAt(time.Now().Add(time.Hour*24)) - p.freeAt(time.Now())
}

func (p *payroll) claim(amount int64) uint64 {
	available := int64(p.available())
	if amount > available || amount < 0 {
		amount = available
	}
	p.totalWithdrawn += uint64(amount)
	return uint64(amount)
}

func (p *payroll) pause(source std.Address) {
	p.assertNotStopped()
	p.assertBeneficiary(source)

	if p.paused {
		panic(errors.New("already paused"))
	}
	p.paused = true
	p.pausedAt = time.Now()
}

func (p *payroll) resume(source std.Address) {
	p.assertNotStopped()
	p.assertBeneficiary(source)

	if !p.paused {
		panic(errors.New("not paused"))
	}
	p.resumeUnsafe()
}

func (p *payroll) resumeUnsafe() {
	p.paused = false
	p.pauseDuration += time.Now().Sub(p.pausedAt)
}

func (p *payroll) stop(source std.Address) {
	p.assertNotStopped()

	var breakupSource BreakupSource
	switch source {
	case p.beneficiary:
		breakupSource = BreakupSourceBeneficiary
	case p.creatorAddr:
		breakupSource = BreakupSourceCreator
	default:
		panic(errors.New("unknown source"))
	}

	if p.paused {
		p.resumeUnsafe()
	}

	p.breakupAmount = p.breakup(time.Now().Sub(p.createdAt), p.pauseDuration, breakupSource)
	p.stopped = true
	p.stoppedAt = time.Now()
}

func (p *payroll) assertNotStopped() {
	if p.stopped {
		panic(errors.New("stopped"))
	}
}

func (p *payroll) assertBeneficiary(addr std.Address) {
	if addr != p.beneficiary {
		panic(errors.New("only beneficiary allowed"))
	}
}

func mustGetPayroll(id seqid.ID) *payroll {
	ip, ok := payrolls.Get(id.String())
	if !ok {
		panic(errors.New("no such payroll"))
	}
	return ip.(*payroll)
}

func DistribStep(amount uint64, per time.Duration) DistribFn {
	return func(elapsed time.Duration) uint64 {
		blocksCount := elapsed / per
		return amount * uint64(blocksCount)
	}
}

func DistribLinear(amount uint64, per time.Duration) DistribFn {
	return func(elapsed time.Duration) uint64 {
		return (uint64(elapsed.Seconds()) * amount) / uint64(per.Seconds())
	}
}

func DistribMonthlyStep(amountPerMonth uint64) DistribFn {
	return DistribStep(amountPerMonth, time.Hour*24*30)
}

func DistribMonthlyContinuous(amountPerMonth uint64) DistribFn {
	return DistribLinear(amountPerMonth, time.Hour*24*30)
}

// CreateCDI creates a payroll that tries to match the French CDI contract
//
// See: https://travail-emploi.gouv.fr/le-contrat-de-travail-duree-indeterminee-cdi
func CreateCDI(namespace string, label string, beneficiary std.Address, amountPerMonth uint64) seqid.ID {
	if amountPerMonth <= 0 {
		panic(errors.New("invalid amount"))
	}
	return Create(namespace, label, beneficiary, -1, DistribMonthlyContinuous(amountPerMonth), BreakupCDI(amountPerMonth))
}

func BreakupCDI(amountPerMonth uint64) BreakupFn {
	const hoursPerYear = float64(8_670)
	return func(elapsed time.Duration, pauseDuration time.Duration, source BreakupSource) int64 {
		if source != BreakupSourceCreator {
			return 0
		}

		// conventional breakup
		worked := elapsed - pauseDuration
		numYears := worked.Hours() / hoursPerYear
		return int64(0.25 * float64(amountPerMonth) * numYears)
	}
}
