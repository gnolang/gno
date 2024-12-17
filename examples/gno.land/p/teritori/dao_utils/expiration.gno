package dao_utils

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/teritori/jsonutil"
	"gno.land/p/demo/json"
)

// loosely ported from https://github.com/CosmWasm/cw-utils/blob/7fce8a214f2f1e7763b8718dcbd2a6dd07f30988/src/expiration.rs

type (
	Expiration interface {
		IsExpired() bool
		ToJSON() *json.Node
		String() string
	}
	ExpirationAtHeight int64
	ExpirationAtTime   time.Time
	ExpirationNever    struct{}
)

func (e ExpirationAtHeight) IsExpired() bool {
	return std.GetHeight() >= int64(e)
}

func (e ExpirationAtHeight) ToJSON() *json.Node {
	val := jsonutil.Int64Node(int64(e))
	return jsonutil.UnionNode("atHeight", val)
}

func (e ExpirationAtHeight) String() string {
	return strconv.FormatInt(int64(e), 10)
}

func (e ExpirationAtTime) IsExpired() bool {
	t := time.Time(e)
	now := time.Now()
	return now.Equal(t) || now.After(t)
}

func (e ExpirationAtTime) ToJSON() *json.Node {
	val := jsonutil.TimeNode(time.Time(e))
	return jsonutil.UnionNode("atTime", val)
}

func (e ExpirationAtTime) String() string {
	return time.Time(e).String()
}

func (e ExpirationNever) IsExpired() bool {
	return false
}

func (e ExpirationNever) ToJSON() *json.Node {
	return jsonutil.UnionNode("never", jsonutil.EmptyObjectNode())
}

func (e ExpirationNever) String() string {
	return "Never"
}

type (
	Duration interface {
		AfterCurrentBlock() Expiration
		ToJSON() *json.Node
		String() string
	}
	DurationHeight int64
	DurationTime   time.Duration
)

func (d DurationHeight) AfterCurrentBlock() Expiration {
	return ExpirationAtHeight(std.GetHeight() + int64(d))
}

func (d DurationHeight) ToJSON() *json.Node {
	val := json.NumberNode("", float64(d))
	return jsonutil.UnionNode("height", val)
}

func (d DurationHeight) String() string {
	return strconv.FormatInt(int64(d), 10)
}

func (d DurationTime) AfterCurrentBlock() Expiration {
	return ExpirationAtTime(time.Now().Add(time.Duration(d)))
}

func (d DurationTime) ToJSON() *json.Node {
	val := jsonutil.DurationNode(time.Duration(d))
	return jsonutil.UnionNode("time", val)
}

func (d DurationTime) String() string {
	return time.Duration(d).String()
}
