package payrolls

import (
	"chain"
	"errors"
	"time"

	"gno.land/p/nt/seqid"
)

type payroll struct {
	id             seqid.ID
	creatorAddr    address
	vaultID        string
	createdAt      time.Time
	beneficiary    address
	totalWithdrawn chain.Coins
	distrib        DistribFn
	breakup        BreakupFn
	paused         bool
	pausedAt       time.Time
	pauseDuration  time.Duration
	label          string
	breakupCoins   chain.Coins
	stopped        bool
	stoppedAt      time.Time
}

type CallSource uint

const (
	CallSourceCreator CallSource = iota
	CallSourceBeneficiary
)

func (p *payroll) freeAt(when time.Time, estimate bool) chain.Coins {
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
	res := p.distrib(cross, elapsed)

	return res
}

func (p *payroll) available() chain.Coins {
	return subCoins(addCoins(p.freeAt(time.Now(), false), p.breakupCoins), p.totalWithdrawn)
}

func (p *payroll) estimateDay() chain.Coins {
	return subCoins(p.freeAt(time.Now().Add(time.Hour*24), true), p.freeAt(time.Now(), true))
}

func (p *payroll) claim(vaultAmount chain.Coins) chain.Coins {
	available := p.available()
	claimable := clampCoins(available, vaultAmount)
	p.totalWithdrawn = addCoins(p.totalWithdrawn, claimable)
	return claimable
}

func (p *payroll) pause(source address) {
	p.assertNotStopped()
	p.assertBeneficiary(source)

	if p.paused {
		panic(errors.New("already paused"))
	}
	p.paused = true
	p.pausedAt = time.Now()
}

func (p *payroll) resume(source address) {
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

func (p *payroll) stop(source address) {
	p.assertNotStopped()

	breakupSource := p.assertSource(source)

	if p.paused {
		p.resumeUnsafe()
	}
	p.breakupCoins = p.getBreakupCoins(breakupSource)
	p.stopped = true
	p.stoppedAt = time.Now()
}

func (p *payroll) getBreakupCoins(source CallSource) chain.Coins {
	if p.breakup == nil {
		return nil
	}

	return p.breakup(cross, time.Now().Sub(p.createdAt), p.pauseDuration, source)
}

func (p *payroll) assertSource(addr address) CallSource {
	switch addr {
	case p.beneficiary:
		return CallSourceBeneficiary
	case p.creatorAddr:
		return CallSourceCreator
	default:
		panic(errors.New("unknown source"))
	}
}

func (p *payroll) assertNotStopped() {
	if p.stopped {
		panic(errors.New("stopped"))
	}
}

func (p *payroll) assertBeneficiary(addr address) {
	if addr != p.beneficiary {
		panic(errors.New("only beneficiary allowed"))
	}
}
