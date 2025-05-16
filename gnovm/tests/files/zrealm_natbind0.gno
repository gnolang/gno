// PKGPATH: gno.land/r/test
package test

import (
	"std"
)

var node any

func init() {
	node = std.ChainHeight
}

func main() {
	crossing()

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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3]=
//     @@ -16,7 +16,7 @@
//                          "Tag": "",
//                          "Type": {
//                              "@type": "/gno.PrimitiveType",
//     -                        "value": "1024"
//     +                        "value": "16"
//                          }
//                      }
//                  ]
//     @@ -24,7 +24,7 @@
//              "V": {
//                  "@type": "/gno.RefValue",
//                  "Escaped": true,
//     -            "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:36"
//     +            "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:33"
//              }
//          }
//      }
// u[a7f5397443359ea76c50be82c77f1f893a060925:33]=
//     @@ -8,9 +8,10 @@
//          "NativePkg": "std",
//          "ObjectInfo": {
//              "ID": "a7f5397443359ea76c50be82c77f1f893a060925:33",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "6",
//              "OwnerID": "a7f5397443359ea76c50be82c77f1f893a060925:2",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// u[a7f5397443359ea76c50be82c77f1f893a060925:36]=
//     @@ -11,7 +11,7 @@
//              "IsEscaped": true,
//              "ModTime": "6",
//              "OwnerID": "a7f5397443359ea76c50be82c77f1f893a060925:2",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
