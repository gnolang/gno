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
	switchrealm()

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
// switchrealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3]=
//     @@ -17,7 +17,7 @@
//                          "Tag": "",
//                          "Type": {
//                              "@type": "/gno.PrimitiveType",
//     -                        "value": "1024"
//     +                        "value": "16"
//                          }
//                      }
//                  ]
//     @@ -31,8 +31,8 @@
//                  },
//                  "FileName": "native.gno",
//                  "IsMethod": false,
//     -            "Name": "ChainHeight",
//     -            "NativeName": "ChainHeight",
//     +            "Name": "ChainID",
//     +            "NativeName": "ChainID",
//                  "NativePkg": "std",
//                  "PkgPath": "std",
//                  "Source": {
//     @@ -41,7 +41,7 @@
//                      "Location": {
//                          "Column": "1",
//                          "File": "native.gno",
//     -                    "Line": "11",
//     +                    "Line": "9",
//                          "PkgPath": "std"
//                      }
//                  },
//     @@ -57,7 +57,7 @@
//                              "Tag": "",
//                              "Type": {
//                                  "@type": "/gno.PrimitiveType",
//     -                            "value": "1024"
//     +                            "value": "16"
//                              }
//                          }
//                      ]
// u[a7f5397443359ea76c50be82c77f1f893a060925:10]=(noop)
