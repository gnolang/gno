package main

import "chain/params"

func main() {
	params.SetString("foo.string", "hello")
	params.SetInt64("bar.int64", -12345)
	params.SetUint64("baz.uint64", 12345)
	params.SetBool("oof.bool", true)
	params.SetBytes("rab.bytes", []byte{0, 1, 2, 255})
	params.SetStrings("strings", []string{"foo", "bar"})

	str, ok := params.GetString("foo.string")
	check("string", ok && str == "hello")

	i64, ok := params.GetInt64("bar.int64")
	check("int64", ok && i64 == -12345)

	u64, ok := params.GetUint64("baz.uint64")
	check("uint64", ok && u64 == 12345)

	b, ok := params.GetBool("oof.bool")
	check("bool", ok && b)

	bz, ok := params.GetBytes("rab.bytes")
	check("bytes", ok && len(bz) == 4 && bz[0] == 0 && bz[1] == 1 && bz[2] == 2 && bz[3] == 255)

	strs, ok := params.GetStrings("strings")
	check("strings", ok && len(strs) == 2 && strs[0] == "foo" && strs[1] == "bar")

	missing, ok := params.GetString("missing")
	check("missing string", !ok && missing == "")

	missingBool, ok := params.GetBool("missing.bool")
	check("missing bool", !ok && !missingBool)

	missingInt64, ok := params.GetInt64("missing.int64")
	check("missing int64", !ok && missingInt64 == 0)

	missingUint64, ok := params.GetUint64("missing.uint64")
	check("missing uint64", !ok && missingUint64 == 0)

	missingBytes, ok := params.GetBytes("missing.bytes")
	check("missing bytes", !ok && len(missingBytes) == 0)

	missingStrings, ok := params.GetStrings("missing.strings")
	check("missing strings", !ok && len(missingStrings) == 0)

	println("done")
}

func check(name string, ok bool) {
	if !ok {
		panic("check failed: " + name)
	}
}

// Output:
// done
