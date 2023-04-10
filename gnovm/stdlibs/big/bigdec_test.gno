package conversion

import (
	"testing"
)

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

// bigdec -> int/int32/int64/uint/uint32/uint64/bigint/string
func TestFromBigdec(t* testing.T) {
	shouldEqual(t, int(mbd), mi)
	shouldEqual(t, int32(mbd), mi32)
	shouldEqual(t, int64(mbd), mi64)
	shouldEqual(t, uint(mbd), mui)
	shouldEqual(t, uint32(mbd), mui32)
	shouldEqual(t, uint64(mbd), mui64)
	shouldEqual(t, bigint(mbd), mbi)
	shouldEqual(t, string(mbd), mstr)
}

