// Package relayer moves drand evmnet beacons onto gno.land for the rounds
// that r/drand has pending requests for.
//
// It is stateless: every tick asks the realm which rounds are pending,
// fetches the ones drand has already published, verifies them locally (so a
// broken mirror never costs gas) and submits them in one transaction. Any
// number of relayers can run side by side; the realm treats a duplicate
// submit as a no-op and verifies every signature, so a relayer is trusted
// for liveness only, never for correctness.
package relayer

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/gnolang/gno/contribs/gnodrand/internal/evmnet"
)

// Beacon is one drand round as served by the drand HTTP API.
type Beacon struct {
	Round     uint64 `json:"round"`
	Signature string `json:"signature"`
}

// Source fetches published beacons.
type Source interface {
	Beacon(ctx context.Context, round uint64) (Beacon, error)
}

// Chain reads pending rounds from, and submits beacons to, r/drand.
type Chain interface {
	PendingRounds(ctx context.Context, limit int) ([]uint64, error)
	Submit(ctx context.Context, beacons []Beacon) error
}

// Relayer is one relay loop.
type Relayer struct {
	Source   Source
	Chain    Chain
	Log      *slog.Logger
	MaxBatch int              // beacons per transaction
	Now      func() time.Time // for tests; defaults to time.Now
}

// Run ticks every interval until ctx is cancelled. Tick errors are logged,
// not fatal: the next tick retries from the chain's state.
func (r *Relayer) Run(ctx context.Context, interval time.Duration) error {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		if n, err := r.Tick(ctx); err != nil {
			r.Log.Error("tick failed", "err", err)
		} else if n > 0 {
			r.Log.Info("submitted beacons", "count", n)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
}

// Tick relays every pending round that drand has already published and
// returns how many beacons it submitted.
func (r *Relayer) Tick(ctx context.Context) (int, error) {
	rounds, err := r.Chain.PendingRounds(ctx, r.MaxBatch)
	if err != nil {
		return 0, fmt.Errorf("pending rounds: %w", err)
	}
	published := evmnet.RoundAt(r.now().Unix())

	var batch []Beacon
	for _, round := range rounds {
		if round > published {
			break // ascending order: the rest is in the future too
		}
		b, err := r.Source.Beacon(ctx, round)
		if err != nil {
			r.Log.Warn("fetch beacon", "round", round, "err", err)
			continue
		}
		if err := check(round, b); err != nil {
			r.Log.Error("rejecting beacon", "round", round, "err", err)
			continue
		}
		batch = append(batch, b)
	}
	if len(batch) == 0 {
		return 0, nil
	}
	if err := r.Chain.Submit(ctx, batch); err != nil {
		return 0, fmt.Errorf("submit %d beacons: %w", len(batch), err)
	}
	return len(batch), nil
}

func (r *Relayer) now() time.Time {
	if r.Now != nil {
		return r.Now()
	}
	return time.Now()
}

func check(round uint64, b Beacon) error {
	if b.Round != round {
		return fmt.Errorf("source returned round %d", b.Round)
	}
	sig, err := hex.DecodeString(b.Signature)
	if err != nil {
		return fmt.Errorf("signature is not hex: %w", err)
	}
	return evmnet.Verify(round, sig)
}
