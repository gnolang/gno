package main

var (
	mi64  int64  = 123
	mui64 uint64 = 123

	mbi bigint = 123

	max_uint64_plus_1 bigint = 18446744073709551616
	max_int64_plus_1  bigint = 9223372036854775808
)

func main() {
	// default value
	var _zeroBig bigint
	shouldEQ(_zeroBig, bigint(0))

	// int64 > bigint
	shouldEQ(bigint(mi64), mbi)

	// uint64 > bigint
	shouldEQ(bigint(mui64), mbi)

	// bigint > int64
	shouldEQ(int64(mbi), mi64)

	// bigint(too big) > int64
	// int64(max_int64_plus_1) // cannot convert BigintKind to Int64Kin

	// bigint > uint64
	shouldEQ(uint64(mbi), mui64)

	// bigint(too big) > uint64
	// uint64(max_uint64_plus_1) // cannot convert BigintKind to Int64Kind

	// OpInc, OpDec
	var _oneBig bigint = 1
	_oneBig++
	shouldEQ(_oneBig, bigint(2))

	_oneBig--
	shouldEQ(_oneBig, bigint(1))
}

func shouldEQ(a, b interface{}) bool {
	if a == b {
		println("true")
	} else {
		println("false")
	}
	return a == b
}

// Output:
// true
// true
// true
// true
// true
// true
// true
