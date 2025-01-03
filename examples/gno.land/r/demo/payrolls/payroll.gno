package payrolls

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/seqid"
)

type payroll struct {
	id             seqid.ID
	creatorAddr    std.Address
	vaultID        string
	createdAt      time.Time
	beneficiary    std.Address
	totalWithdrawn std.Coins
	distrib        DistribFn
	breakup        BreakupFn
	paused         bool
	pausedAt       time.Time
	pauseDuration  time.Duration
	label          string
	breakupCoins   std.Coins
	stopped        bool
	stoppedAt      time.Time
}

func (p *payroll) freeAt(when time.Time, estimate bool) std.Coins {
	if p.stopped && p.stoppedAt.Before(when) {
		when = p.stoppedAt
	}

	var elapsed time.Duration
	if p.paused && !estimate {
		elapsed = p.pausedAt.Sub(p.createdAt)
	} else {
		elapsed = when.Sub(p.createdAt)
	}
	elapsed -= p.pauseDuration
	res := p.distrib(elapsed)

	return res
}

func (p *payroll) available() std.Coins {
	return subCoins(addCoins(p.freeAt(time.Now(), false), p.breakupCoins), p.totalWithdrawn)
}

func (p *payroll) estimateDay() std.Coins {
	return subCoins(p.freeAt(time.Now().Add(time.Hour*24), true), p.freeAt(time.Now(), true))
}

func (p *payroll) claim(vaultAmount std.Coins) std.Coins {
	available := p.available()
	claimable := clampCoins(available, vaultAmount)
	p.totalWithdrawn = addCoins(p.totalWithdrawn, claimable)
	return claimable
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

	if p.paused {
		p.resumeUnsafe()
	}

	p.breakupCoins = p.getBreakupCoins(source)
	p.stopped = true
	p.stoppedAt = time.Now()
}

func (p *payroll) getBreakupCoins(source std.Address) std.Coins {
	var breakupSource BreakupSource
	switch source {
	case p.beneficiary:
		breakupSource = BreakupSourceBeneficiary
	case p.creatorAddr:
		breakupSource = BreakupSourceCreator
	default:
		panic(errors.New("unknown source"))
	}

	return p.breakup(time.Now().Sub(p.createdAt), p.pauseDuration, breakupSource)
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
