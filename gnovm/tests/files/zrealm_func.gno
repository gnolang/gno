// PKGPATH: gno.land/r/test
package test

var a any

func init() {
	a = func(a int) {
	}
}

func main(cur realm) {
	a = func(b string) {
		println("dontcare")
	}
	println("done")
}

// Output:
// done

// Realm:
// finalizerealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:8](309)={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "LastObjectSize": "309",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/test",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "zrealm_func.gno",
//             "PkgPath": "gno.land/r/test",
//             "Span": {
//                 "End": {
//                     "Column": "3",
//                     "Line": "14"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "6",
//                     "Line": "12"
//                 }
//             }
//         }
//     },
//     "Type": {
//         "@type": "/gno.FuncType",
//         "Params": [
//             {
//                 "Embedded": false,
//                 "Name": "b",
//                 "Tag": "",
//                 "Type": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "16"
//                 }
//             }
//         ],
//         "Results": []
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "351",
//     -        "ModTime": "6",
//     +        "ModTime": "7",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -12,11 +12,11 @@
//                  "Params": [
//                      {
//                          "Embedded": false,
//     -                    "Name": "a",
//     +                    "Name": "b",
//                          "Tag": "",
//                          "Type": {
//                              "@type": "/gno.PrimitiveType",
//     -                        "value": "32"
//     +                        "value": "16"
//                          }
//                      }
//                  ],
//     @@ -24,8 +24,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "8806871058538d9ab7d4c168c6788e300f7e72fa",
//     -            "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +            "Hash": "7b5d94f791350d9cbeab878b065cdf86f258c507",
//     +            "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//              }
//          }
//      }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](-309)
