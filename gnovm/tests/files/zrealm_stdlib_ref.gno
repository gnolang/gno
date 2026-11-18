// PKGPATH: gno.land/r/test
package test

import "strconv"

// Store a reference to an existing stdlib object (not a copy).
var ref interface{}

func main(cur realm) {
	// This should store a reference to strconv.Itoa itself,
	// which is a FuncValue owned by the strconv package.
	// Does this incref the stdlib object?
	ref = strconv.Itoa
	println(ref.(func(int) string)(42))
}

// Output:
// 42

// Realm:
// finalizerealm["gno.land/r/test"]
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:6](338)={
//     "Crossing": false,
//     "FileName": "itoa.gno",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "Itoa",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "LastObjectSize": "338",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "strconv",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "itoa.gno",
//             "PkgPath": "strconv",
//             "Span": {
//                 "End": {
//                     "Column": "2",
//                     "Line": "36"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "1",
//                     "Line": "34"
//                 }
//             }
//         }
//     },
//     "Type": {
//         "@type": "/gno.FuncType",
//         "Params": [
//             {
//                 "Embedded": false,
//                 "Name": "i",
//                 "PkgPath": "",
//                 "Tag": "",
//                 "Type": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "32"
//                 }
//             }
//         ],
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
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](197)=
//     @@ -2,9 +2,42 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.FuncType",
//     +            "Params": [
//     +                {
//     +                    "Embedded": false,
//     +                    "Name": "i",
//     +                    "PkgPath": "",
//     +                    "Tag": "",
//     +                    "Type": {
//     +                        "@type": "/gno.PrimitiveType",
//     +                        "value": "32"
//     +                    }
//     +                }
//     +            ],
//     +            "Results": [
//     +                {
//     +                    "Embedded": false,
//     +                    "Name": ".res.0",
//     +                    "PkgPath": "",
//     +                    "Tag": "",
//     +                    "Type": {
//     +                        "@type": "/gno.PrimitiveType",
//     +                        "value": "16"
//     +                    }
//     +                }
//     +            ]
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "befdf02ecfb9b53bf7cb7a9cefed2a26321b2f59",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6"
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "537",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "77f76390147dcb12d9de104eb549c3a959934c1c",
//     +                "Hash": "e3a1e645a395857719c74263f68b24d1977ef3e5",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
