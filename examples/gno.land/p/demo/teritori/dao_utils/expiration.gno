package dao_utils

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/teritori/ujson"
)

// loosely ported from https://github.com/CosmWasm/cw-utils/blob/7fce8a214f2f1e7763b8718dcbd2a6dd07f30988/src/expiration.rs

type (
	Expiration interface {
		IsExpired() bool
		ToJSON() string
		String() string
	}
	ExpirationAtHeight int64
	ExpirationAtTime   time.Time
	ExpirationNever    struct{}
)

func (e ExpirationAtHeight) IsExpired() bool {
	return std.GetHeight() >= int64(e)
}

func (e ExpirationAtHeight) ToJSON() string {
	return ujson.FormatUnionMember("atHeight", int64(e), false)
}

func (e ExpirationAtHeight) String() string {
	return strconv.FormatInt(int64(e), 10)
}

func (e ExpirationAtTime) IsExpired() bool {
	return time.Now().After(time.Time(e))
}

func (e ExpirationAtTime) ToJSON() string {
	return ujson.FormatUnionMember("atTime", time.Time(e), false)
}

func (e ExpirationAtTime) String() string {
	return time.Time(e).String()
}

func (e ExpirationNever) IsExpired() bool {
	return false
}

func (e ExpirationNever) ToJSON() string {
	return ujson.FormatUnionMember("never", "{}", true)
}

func (e ExpirationNever) String() string {
	return "Never"
}

type (
	Duration interface {
		AfterCurrentBlock() Expiration
		ToJSON() string
		String() string
	}
	DurationHeight int64
	DurationTime   time.Duration
)

func (d DurationHeight) AfterCurrentBlock() Expiration {
	return ExpirationAtHeight(std.GetHeight() + int64(d))
}

func (d DurationHeight) ToJSON() string {
	return ujson.FormatUnionMember("height", int64(d), false)
}

func (d DurationHeight) String() string {
	return strconv.FormatInt(int64(d), 10)
}

func (d DurationTime) AfterCurrentBlock() Expiration {
	return ExpirationAtTime(time.Now().Add(time.Duration(d)))
}

func (d DurationTime) ToJSON() string {
	return ujson.FormatUnionMember("time", time.Duration(d), false)
}

func (d DurationTime) String() string {
	return time.Duration(d).String()
}
