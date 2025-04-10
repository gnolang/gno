// PKGPATH: gno.land/r/test
package test

import "chain/runtime"

var node any

func init() {
	node = runtime.ChainHeight
}

func main() {
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
// switchrealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2]=
//     @@ -30,7 +30,7 @@
//                              "Tag": "",
//                              "Type": {
//                                  "@type": "/gno.PrimitiveType",
//     -                            "value": "1024"
//     +                            "value": "16"
//                              }
//                          }
//                      ]
//     @@ -44,8 +44,8 @@
//                      },
//                      "FileName": "native.gno",
//                      "IsMethod": false,
//     -                "Name": "ChainHeight",
//     -                "NativeName": "ChainHeight",
//     +                "Name": "ChainID",
//     +                "NativeName": "ChainID",
//                      "NativePkg": "chain/runtime",
//                      "PkgPath": "chain/runtime",
//                      "Source": {
//     @@ -54,7 +54,7 @@
//                          "Location": {
//                              "Column": "1",
//                              "File": "native.gno",
//     -                        "Line": "15",
//     +                        "Line": "13",
//                              "PkgPath": "chain/runtime"
//                          }
//                      },
//     @@ -69,7 +69,7 @@
//                                  "Tag": "",
//                                  "Type": {
//                                      "@type": "/gno.PrimitiveType",
//     -                                "value": "1024"
//     +                                "value": "16"
//                                  }
//                              }
//                          ]
// u[c09c8277a76bf0c457fdf56bd592edcdcf839a50:5]=(noop)
