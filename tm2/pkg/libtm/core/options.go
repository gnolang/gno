package core

import "log/slog"

type Option func(t *Tendermint)

// WithLogger specifies the logger for the Tendermint consensus engine
func WithLogger(l *slog.Logger) Option {
	return func(t *Tendermint) {
		t.logger = l
	}
}

// WithProposeTimeout specifies the propose state timeout
func WithProposeTimeout(timeout Timeout) Option {
	return func(t *Tendermint) {
		t.timeouts[propose] = timeout
	}
}

// WithPrevoteTimeout specifies the prevote state timeout
func WithPrevoteTimeout(timeout Timeout) Option {
	return func(t *Tendermint) {
		t.timeouts[prevote] = timeout
	}
}

// WithPrecommitTimeout specifies the precommit state timeout
func WithPrecommitTimeout(timeout Timeout) Option {
	return func(t *Tendermint) {
		t.timeouts[precommit] = timeout
	}
}
