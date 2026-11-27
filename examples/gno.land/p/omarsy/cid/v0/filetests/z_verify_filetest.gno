package main

import cid "gno.land/p/omarsy/cid/v0"

func main() {
	c, err := cid.Parse("bafkreibm6jg3ux5qumhcn2b3flc3tyu6dmlb4xa7u5bf44yegnrjhc4yeq")
	if err != nil {
		panic(err)
	}

	println(c.Version(), c.Codec() == cid.Raw)
	println(c.Verify([]byte("hello")) == nil)
	println(c.Verify([]byte("hellp")) == cid.ErrDigestMismatch)
	println(cid.Sum(cid.Raw, []byte("hello")) == c)
}

// Output:
// 1 true
// true
// true
// true
