// PKGPATH: gno.land/r/test
package test

import "chain/runtime"

var node any

func init() {
	node = runtime.ChainHeight
}

func main(cur realm) {
	// NOTE: this test uses ChainHeight and ChainID, which are "pure"
	// natively bound functions (ie. not indirections through a wrapper fn,
	// to convert the types to builtin go/gno identifiers).
	f := node.(func() int64)
	println(f())
	node = runtime.ChainID
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
//     -            "ObjectID": "c09c8277a76bf0c457fdf56bd592edcdcf839a50:8"
//     +            "ObjectID": "c09c8277a76bf0c457fdf56bd592edcdcf839a50:5"
//              }
//          }
//      }
// u[c09c8277a76bf0c457fdf56bd592edcdcf839a50:5](7)=
//     @@ -8,10 +8,11 @@
//          "NativePkg": "chain/runtime",
//          "ObjectInfo": {
//              "ID": "c09c8277a76bf0c457fdf56bd592edcdcf839a50:5",
//     +        "IsEscaped": true,
//              "LastObjectSize": "457",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "c09c8277a76bf0c457fdf56bd592edcdcf839a50:2",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// u[c09c8277a76bf0c457fdf56bd592edcdcf839a50:8](0)=
//     @@ -12,7 +12,7 @@
//              "LastObjectSize": "473",
//              "ModTime": "6",
//              "OwnerID": "c09c8277a76bf0c457fdf56bd592edcdcf839a50:2",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
