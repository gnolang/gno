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
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](-2)=
//     @@ -18,15 +18,15 @@
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
//     -            "Hash": "34cd2ba539cb8490b2304a329a3f44edd89d0425",
//     -            "ObjectID": "c09c8277a76bf0c457fdf56bd592edcdcf839a50:10"
//     +            "Hash": "979adacee58d2714ae2a70c8d2e6b38d715321dc",
//     +            "ObjectID": "c09c8277a76bf0c457fdf56bd592edcdcf839a50:7"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "aa1778bedd085a52f08a1834556286316a3f97cf",
//     +                "Hash": "d1bc5077a7e598bcefa531b2e3d30812765544be",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
