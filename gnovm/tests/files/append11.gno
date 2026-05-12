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
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4](0)=
//     @@ -1,5 +1,5 @@
//      {
//     -    "Data": "YWJjAAAAAAAAAA==",
//     +    "Data": "YWJjZAAAAAAAAA==",
//          "List": null,
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3](0)=
//     @@ -19,7 +19,7 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "a4a1236922e5802c4169a2cb9c125ee72c5e32ff",
//     +                "Hash": "b4ab4a7fafb89dd35cd913f4021ef7b8b47b34f0",
//                      "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4"
//                  },
//                  "Length": "0",
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "6db85c90eb7700d0409b20daefc92fdd5853fe84",
//     +                "Hash": "6294b6b4e419e83eb5743111b93b9ee3ebb27255",
//                      "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3"
//                  }
//              },
