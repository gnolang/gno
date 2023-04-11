package conversion

import (
	"testing"
)

var (
	mi 		int			= 123
	mi32	int32 	= 123
	mi64	int64 	= 123
	
	mui		uint		= 123
	mui32	uint32	= 123
	mui64	uint64	= 123

	mbi		bigint	= 123
	mbd		bigdec	= 123
	mstr	string	= "123"
)

// bigint -> bigdec/string 
func TestFromBigint(t* testing.T) {
	shouldEqual(t, bigdec(mbi), mbd)
	shouldEqual(t, string(mbi), mstr)
}

// int/int32/int64/uint/uint32/uint64/bigdec/string -> bigint
func TestToBigint(t *testing.T) {
	shouldEqual(t, bigint(mi), mbi)
	shouldEqual(t, bigint(mi32), mbi)
	shouldEqual(t, bigint(mi64), mbi)
	shouldEqual(t, bigint(mui), mbi)
	shouldEqual(t, bigint(mui32), mbi)
	shouldEqual(t, bigint(mui64), mbi)
	shouldEqual(t, bigint(mstr), mbi)
}

func TestDefaultBigint(t *testing.T) {
	var _bi bigint
	shouldEqual(t, _bi, bigint(0))
}