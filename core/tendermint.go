package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"
)

type State int

const (
	Propose State = iota
	Prevote
	Precommit
)

type Proposer interface {
	Get(height uint64, block []byte, round int64) []byte
	Me() []byte
}

type Broadcast interface {
	Broadcast(state State, height uint64, round int64, block []byte)
}

type BlockValidator interface {
	IsValid(height uint64, block []byte) bool
	IsMajority(votes int) bool
}

// TODO define the finalized proposal

// `NewTendermint` and `Init` must be called,
// in that order. before any messages can be processed
type Tendermint struct {
	state        State
	round        int64
	lockedBlock  []byte
	lockedRound  int64
	validBlock   []byte
	validRound   int64
	init         bool
	cfg          *Config
	validMsgsCnt [Precommit + 1]int
	done         bool
}

type Config struct {
	ctx     context.Context
	height  uint64
	p       Proposer
	b       Broadcast
	timeout time.Duration
	block   []byte
	bv      BlockValidator
}

func NewTendermint(cfg *Config) *Tendermint {
	t := &Tendermint{
		state:        Propose,
		round:        0,
		lockedBlock:  nil,
		lockedRound:  -1,
		validBlock:   nil,
		validRound:   -1,
		cfg:          cfg,
		validMsgsCnt: [Precommit + 1]int{0, 0, 0},
	}

	return t
}

type Msg struct {
	state  State
	height uint64
	round  int64
	block  []byte
	from   []byte
}

func (t *Tendermint) Init() error {
	if !t.cfg.bv.IsValid(t.cfg.height, t.cfg.block) {
		return errors.New(fmt.Sprintf("invalid height: %v block: %v\n", t.cfg.height, t.cfg.block))
	}

	t.init = true
	proposer := t.cfg.p.Get(t.cfg.height, t.cfg.block, t.round)

	if bytes.Equal(t.cfg.p.Me(), proposer) {
		t.cfg.b.Broadcast(Propose, t.cfg.height, t.round, t.cfg.block)
	} else {
		time.Sleep(t.cfg.timeout)
		t.cfg.b.Broadcast(Prevote, t.cfg.height, t.round, t.cfg.block)
		t.state = Prevote
	}
	return nil
}

// ProcessMsg accepts a *Msg and potentially advances the state
// of the consensus. It returns an error if there was such and a boolean
// value indicating whether consensus has been reached or not.
func (t *Tendermint) ProcessMsg(m *Msg) (error, bool) {
	if !t.cfg.bv.IsValid(m.height, m.block) {
		return errors.New(fmt.Sprintf("invalid height: %v block: %v\n", m.height, m.block)), false
	}
	if !t.init {
		return errors.New("not initialized"), false
	}

	if m.height != t.cfg.height {
		return errors.New(fmt.Sprintf("expected height: %v got: %v\n", t.cfg.height, m.height)), false
	}

	proposer := t.cfg.p.Get(t.cfg.height, t.cfg.block, t.round)

	switch m.state {
	case Propose:
		if !bytes.Equal(m.from, proposer) {
			return errors.New(fmt.Sprintf("expected proposer: %v got: %v\n", proposer, m.from)), false
		}

		switch {
		//22:
		case t.state == Propose && t.validRound == -1:
			if t.lockedRound == -1 || bytes.Equal(t.lockedBlock, m.block) {
				t.cfg.b.Broadcast(Prevote, m.height, m.round, m.block)
			} else {
				t.cfg.b.Broadcast(Prevote, m.height, m.round, nil)
			}
		//28:
		case t.state == Propose && t.cfg.bv.IsMajority(t.validMsgsCnt[Prevote]):
			if t.validRound >= 0 || t.validRound < t.round {
				if t.lockedRound == -1 || bytes.Equal(t.lockedBlock, m.block) {
					t.cfg.b.Broadcast(Prevote, m.height, m.round, m.block)
				} else {
					t.cfg.b.Broadcast(Prevote, m.height, m.round, nil)
				}
			}
		//36:
		case t.state >= Prevote && t.cfg.bv.IsMajority(t.validMsgsCnt[Prevote]):
			if t.state == Prevote {
				t.lockedBlock = m.block
				t.lockedRound = m.round
				t.cfg.b.Broadcast(Precommit, m.height, m.round, m.block)
				t.state = Precommit
			}
			t.validBlock = m.block
			t.validRound = t.round

		//49:
		case t.state >= Prevote && t.cfg.bv.IsMajority(t.validMsgsCnt[Precommit]):
			decision := true //while decision[t.height] == nil

			if decision {
				//decision[t.height] == m.block
				t.cfg.height += 1
				t.round = 0
				err := t.Init()

				if err != nil {
					return err, false
				}
			}
		}

		t.state = Prevote
		t.validMsgsCnt[Propose] += 1
	case Prevote:
		if t.state == Prevote && t.cfg.bv.IsMajority(t.validMsgsCnt[Prevote]) {
			//44:
			if m.block == nil {
				t.cfg.b.Broadcast(Precommit, t.cfg.height, t.round, nil)
			} else {
				//34:
				time.Sleep(t.cfg.timeout)
				if t.round == m.round && t.state == Prevote {
					t.cfg.b.Broadcast(Precommit, t.cfg.height, t.round, t.cfg.block)
				}
			}
			t.state = Precommit
		}

		t.validMsgsCnt[Prevote] += 1
	//47:
	case Precommit:
		if m.round == t.round && t.cfg.bv.IsMajority(t.validMsgsCnt[Precommit]) {
			var err error

			time.Sleep(t.cfg.timeout)
			if t.round == m.round && t.state == Prevote {
				t.round += 1
				err = t.Init()
			}

			if err != nil {
				return err, false
			}
			t.done = true
		}

		t.validMsgsCnt[Precommit] += 1
	}

	return nil, t.done
}
