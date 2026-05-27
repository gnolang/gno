package main

import "chain/params"

func main() {
	params.SetString("foo.string", "hello")
	params.SetInt64("bar.int64", -12345)
	params.SetUint64("baz.uint64", 12345)
	params.SetBool("oof.bool", true)
	params.SetBytes("rab.bytes", []byte("world"))

	colonKey := []byte{0x01, 0x3a, 0x02}
	colonValue := []byte{0xaa, 0xbb, 0xcc}
	params.SetBytesKey(colonKey, colonValue)
	checkBytes("colon key", mustGetBytesKey(colonKey), colonValue)

	h256Key := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	}
	h256Value := []byte{
		0xff, 0xfe, 0xfd, 0xfc, 0xfb, 0xfa, 0xf9, 0xf8,
		0xf7, 0xf6, 0xf5, 0xf4, 0xf3, 0xf2, 0xf1, 0xf0,
		0xef, 0xee, 0xed, 0xec, 0xeb, 0xea, 0xe9, 0xe8,
		0xe7, 0xe6, 0xe5, 0xe4, 0xe3, 0xe2, 0xe1, 0xe0,
	}
	params.SetBytesKey(h256Key, h256Value)
	checkBytes("h256 key", mustGetBytesKey(h256Key), h256Value)

	missing, ok := params.GetBytesKey([]byte{0x99, 0x3a, 0x01})
	check("missing key", !ok && len(missing) == 0)

	params.SetBytesKey(colonKey, nil)
	deleted, ok := params.GetBytesKey(colonKey)
	check("deleted key", !ok && len(deleted) == 0)

	println("done")
}

func mustGetBytesKey(key []byte) []byte {
	value, ok := params.GetBytesKey(key)
	if !ok {
		panic("missing raw key")
	}
	return value
}

func checkBytes(name string, got []byte, want []byte) {
	check(name+" length", len(got) == len(want))
	for i := 0; i < len(want); i++ {
		check(name+" value", got[i] == want[i])
	}
}

func check(name string, ok bool) {
	if !ok {
		panic("check failed: " + name)
	}
}

// Output:
// done
