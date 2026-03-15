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
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3](0)=
//     @@ -19,7 +19,7 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "927a35f51e8351373cf24a58dfb6105e492cce45",
//     +                "Hash": "2fd96a86edda58aef40cd49ab3c7a3366551c102",
//                      "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4"
//                  },
//                  "Length": "0",
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "290e21aef1ab5f68cd8ea185db45c101c9f03f02",
//     +                "Hash": "9a3dc04f8bd4bcfbddaaffa3379e6bd28afd1fe2",
//                      "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3"
//                  }
//              },
