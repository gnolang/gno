package core

import (
	"std"
	"time"
)

// TODO: use this pattern for Threshold and remove this file

// Ported from https://github.com/CosmWasm/cw-utils/blob/7fce8a214f2f1e7763b8718dcbd2a6dd07f30988/src/expiration.rs

type (
	Expiration interface {
		IsExpired() bool
	}
	ExpirationAtHeight int64
	ExpirationAtTime   time.Time
	ExpirationNever    struct{}
)

func (e ExpirationAtHeight) IsExpired() bool {
	return std.GetHeight() >= int64(e)
}

func (e ExpirationAtTime) IsExpired() bool {
	return time.Now().After(time.Time(e))
}

func (e ExpirationNever) IsExpired() bool {
	return false
}

type (
	Duration interface {
		AfterCurrentBlock() Expiration
	}
	DurationHeight int64
	DurationTime   time.Duration
)

func (d DurationHeight) AfterCurrentBlock() Expiration {
	return ExpirationAtHeight(std.GetHeight() + int64(d))
}

func (d DurationTime) AfterCurrentBlock() Expiration {
	return ExpirationAtTime(time.Now().Add(time.Duration(d)))
}
