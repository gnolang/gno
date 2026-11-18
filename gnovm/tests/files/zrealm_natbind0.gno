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
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:8](350)={
//     "Crossing": false,
//     "FileName": "native.gno",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "ChainID",
//     "NativeName": "ChainID",
//     "NativePkg": "chain/runtime",
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "LastObjectSize": "350",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "chain/runtime",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "native.gno",
//             "PkgPath": "chain/runtime",
//             "Span": {
//                 "End": {
//                     "Column": "22",
//                     "Line": "9"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "1",
//                     "Line": "9"
//                 }
//             }
//         }
//     },
//     "Type": {
//         "@type": "/gno.FuncType",
//         "Params": [],
//         "Results": [
//             {
//                 "Embedded": false,
//                 "Name": ".res.0",
//                 "PkgPath": "",
//                 "Tag": "",
//                 "Type": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "16"
//                 }
//             }
//         ]
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](-1)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "357",
//     -        "ModTime": "6",
//     +        "ModTime": "7",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
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
//     -            "Hash": "02f96dfeaa2a9d0b60f4f966e3c6e49bda3df365",
//     -            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +            "Hash": "5766702f76a24821bb169bd152e36a1ef54d28da",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "677",
//     -        "ModTime": "7",
//     +        "ModTime": "8",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "56176213bef7304614be3fc11403a2766650c774",
//     +                "Hash": "c9f7ad2cadc4abcd78b9dc0de991cba06e3a7ae0",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:7](-359)
