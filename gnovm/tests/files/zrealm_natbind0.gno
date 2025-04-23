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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](-1)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "ModTime": "6",
//     -        "ObjectSize": "190",
//     +        "ObjectSize": "360",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//                          "Tag": "",
//                          "Type": {
//                              "@type": "/gno.PrimitiveType",
//     -                        "value": "1024"
//     +                        "value": "16"
//                          }
//                      }
//                  ]
//     @@ -26,7 +26,7 @@
//              "V": {
//                  "@type": "/gno.RefValue",
//                  "Escaped": true,
//     -            "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:36"
//     +            "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:33"
//              }
//          }
//      }
// u[a7f5397443359ea76c50be82c77f1f893a060925:33](7)=
//     @@ -8,10 +8,11 @@
//          "NativePkg": "std",
//          "ObjectInfo": {
//              "ID": "a7f5397443359ea76c50be82c77f1f893a060925:33",
//     -        "ModTime": "0",
//     -        "ObjectSize": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "6",
//     +        "ObjectSize": "419",
//              "OwnerID": "a7f5397443359ea76c50be82c77f1f893a060925:2",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// u[a7f5397443359ea76c50be82c77f1f893a060925:36](0)=
//     @@ -10,9 +10,9 @@
//              "ID": "a7f5397443359ea76c50be82c77f1f893a060925:36",
//              "IsEscaped": true,
//              "ModTime": "6",
//     -        "ObjectSize": "428",
//     +        "ObjectSize": "435",
//              "OwnerID": "a7f5397443359ea76c50be82c77f1f893a060925:2",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
