// PKGPATH: gno.land/r/main
package main

var bs []byte = make([]byte, 0, 10)

func init() {
	_ = append(bs, byte('a'), byte('b'), byte('c'))
}

func main() {
	println(string(bs[0:3]))
	_ = append(bs, byte('a'), byte('b'), byte('c'), byte('d'))
	println(string(bs[0:4]))
}

// Output:
// abc
// abcd

// Realm:
// finalizerealm["gno.land/r/main"]
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4](0)=
//     @@ -1,5 +1,5 @@
//      {
//     -    "Data": "YWJjAAAAAAAAAA==",
//     +    "Data": "YWJjZAAAAAAAAA==",
//          "List": null,
//          "ObjectInfo": {
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
