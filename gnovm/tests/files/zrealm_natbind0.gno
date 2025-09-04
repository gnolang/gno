// PKGPATH: gno.land/r/test
package test

import (
	"std"
)

var node any

func init() {
	node = std.ChainHeight
}

func main(cur realm) {
	// NOTE: this test uses ChainHeight and ChainID, which are "pure"
	// natively bound functions (ie. not indirections through a wrapper fn,
	// to convert the types to builtin go/gno identifiers).
	f := node.(func() int64)
	println(f())
	node = std.ChainID
	g := node.(func() string)
	println(g())
}

// Output:
// 123
// dev

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](-1)=
//     @@ -17,7 +17,7 @@
//                          "Tag": "",
//                          "Type": {
//                              "@type": "/gno.PrimitiveType",
//     -                        "value": "1024"
//     +                        "value": "16"
//                          }
//                      }
//                  ]
//     @@ -25,7 +25,7 @@
//              "V": {
//                  "@type": "/gno.RefValue",
//                  "Escaped": true,
//     -            "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:42"
//     +            "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:39"
//              }
//          }
//      }
// u[a7f5397443359ea76c50be82c77f1f893a060925:39](7)=
//     @@ -8,10 +8,11 @@
//          "NativePkg": "std",
//          "ObjectInfo": {
//              "ID": "a7f5397443359ea76c50be82c77f1f893a060925:39",
//     +        "IsEscaped": true,
//              "LastObjectSize": "429",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "a7f5397443359ea76c50be82c77f1f893a060925:2",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// u[a7f5397443359ea76c50be82c77f1f893a060925:42](0)=
//     @@ -12,7 +12,7 @@
//              "LastObjectSize": "445",
//              "ModTime": "6",
//              "OwnerID": "a7f5397443359ea76c50be82c77f1f893a060925:2",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
