package types

import (
	"sync/atomic"
)

// DeadlineExceededError is the panic of a deadline gas meter whose deadline has
// passed. It is not an OutOfGasError: the work did not run out of gas, it ran
// longer than the node allows, so a caller tells the two apart by type.
type DeadlineExceededError struct {
	Descriptor string
}

func (e DeadlineExceededError) Error() string {
	return "deadline exceeded in location: " + e.Descriptor
}

// deadlineGasMeter is a basicGasMeter that also stops once a shared flag is
// set, by panicking with DeadlineExceededError from ConsumeGas.
//
// It is NOT deterministic: whether and where it fires depends on the speed of
// the machine and its load. Use it only where the result is not part of
// consensus, such as ABCI queries.
//
// The flag is owned by the caller, typically set by a time.AfterFunc. It is
// sticky, so recovering the panic somewhere cannot buy more time: the next
// charge panics again. The meter embeds the concrete basicGasMeter rather than
// a GasMeter interface, so a charge stays one dynamic call plus one atomic load.
type deadlineGasMeter struct {
	basicGasMeter
	expired *atomic.Bool
}

// NewDeadlineGasMeter returns a gas meter with the given limit that also stops
// once expired is set. A nil expired gives a plain NewGasMeter(limit).
func NewDeadlineGasMeter(limit Gas, expired *atomic.Bool) GasMeter {
	if expired == nil {
		return NewGasMeter(limit)
	}
	return &deadlineGasMeter{basicGasMeter: *NewGasMeter(limit), expired: expired}
}

func (g *deadlineGasMeter) ConsumeGas(amount Gas, descriptor string) {
	// Charge first, so a real out-of-gas is reported as such.
	g.basicGasMeter.ConsumeGas(amount, descriptor)
	if g.expired.Load() {
		panic(DeadlineExceededError{Descriptor: descriptor})
	}
}
