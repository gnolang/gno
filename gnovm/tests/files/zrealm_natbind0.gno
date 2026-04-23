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
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](-1)=
//     @@ -17,15 +17,15 @@
//                          "Tag": "",
//                          "Type": {
//                              "@type": "/gno.PrimitiveType",
//     -                        "value": "1024"
//     +                        "value": "16"
//                          }
//                      }
//                  ]
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "3f7b8524def84bc42c4301dc9b60ea7bb2b8d77b",
//     -            "ObjectID": "c09c8277a76bf0c457fdf56bd592edcdcf839a50:8"
//     +            "Hash": "da89a2b407d7ceb547ae17e162a06607f8e11c06",
//     +            "ObjectID": "c09c8277a76bf0c457fdf56bd592edcdcf839a50:5"
//              }
//          }
//      }
