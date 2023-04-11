package conversion

import (
	"testing"
)

// bigdec -> string
func TestFromBigdec(t* testing.T) {
	shouldEqual(t, string(mbd), mstr)
}

// int/int32/int64/uint/uint32/uint64/bigint/string -> bigdec
func TestToBigdec(t *testing.T) {
	shouldEqual(t, bigdec(mi), mbd)
	shouldEqual(t, bigdec(mi32), mbd)
	shouldEqual(t, bigdec(mi64), mbd)
	shouldEqual(t, bigdec(mui), mbd)
	shouldEqual(t, bigdec(mui32), mbd)
	shouldEqual(t, bigdec(mui64), mbd)
	shouldEqual(t, bigdec(mbd), mbd)
	shouldEqual(t, bigdec(mstr), mbd)

	// should always panic
	// XXX how to catch go panic in gno ??
	// shouldPanic(t, func() {bigdec("abc")} )
}