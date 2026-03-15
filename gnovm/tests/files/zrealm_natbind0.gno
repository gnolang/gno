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
//     @@ -8,10 +8,10 @@
//          "NativePkg": "chain/runtime",
//          "ObjectInfo": {
//              "ID": "c09c8277a76bf0c457fdf56bd592edcdcf839a50:5",
//     +        "IsEscaped": true,
//              "LastObjectSize": "457",
//     -        "ModTime": "0",
//     -        "OwnerID": "c09c8277a76bf0c457fdf56bd592edcdcf839a50:2",
//     -        "RefCount": "1"
//     +        "ModTime": "6",
//     +        "RefCount": "2"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// u[c09c8277a76bf0c457fdf56bd592edcdcf839a50:8](0)=
//     @@ -11,7 +11,7 @@
//              "IsEscaped": true,
//              "LastObjectSize": "473",
//              "ModTime": "6",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// u[c09c8277a76bf0c457fdf56bd592edcdcf839a50:2](2)=
//     @@ -931,7 +931,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "319737b8539492457971c445813bbdf68ae5214a",
//     +                "Escaped": true,
//                      "ObjectID": "c09c8277a76bf0c457fdf56bd592edcdcf839a50:5"
//                  }
//              },
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "b13bfc7e32849fef3e3fcafe63cfbf3d72ab7342",
//     +                "Hash": "22237ce5961c2a62ec15ded22031848f4d83337d",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
