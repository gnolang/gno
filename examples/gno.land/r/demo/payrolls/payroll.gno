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
