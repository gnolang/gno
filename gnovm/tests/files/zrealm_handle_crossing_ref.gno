// PKGPATH: gno.land/r/zrealm_handle_crossing_ref
package zrealm_handle_crossing_ref

// Pins that a crossing function stored as a value stays a reference to the
// package's own function object while a non-crossing one becomes a handle.

var g func(realm)
var p func()

func Gate(cur realm) {}

func Plain() {}

func main(cur realm) {
	g = Gate
	p = Plain
	println("set")
}

// Output:
// set

// Realm:
// finalizerealm["gno.land/r/zrealm_handle_crossing_ref"]
// c[05afb073ed788a666b933765c9bde3fa80fa7991:9](374)={
//     "Crossing": false,
//     "FileName": "zrealm_handle_crossing_ref.gno",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "Plain",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "05afb073ed788a666b933765c9bde3fa80fa7991:9",
//         "LastObjectSize": "374",
//         "ModTime": "0",
//         "OwnerID": "05afb073ed788a666b933765c9bde3fa80fa7991:4",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/zrealm_handle_crossing_ref",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "zrealm_handle_crossing_ref.gno",
//             "PkgPath": "gno.land/r/zrealm_handle_crossing_ref",
//             "Span": {
//                 "End": {
//                     "Column": "16",
//                     "Line": "12"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "1",
//                     "Line": "12"
//                 }
//             }
//         }
//     },
//     "Type": {
//         "@type": "/gno.FuncType",
//         "Params": [],
//         "Results": []
//     }
// }
// u[05afb073ed788a666b933765c9bde3fa80fa7991:3](108)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "05afb073ed788a666b933765c9bde3fa80fa7991:3",
//              "LastObjectSize": "252",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "05afb073ed788a666b933765c9bde3fa80fa7991:2",
//              "RefCount": "1"
//          },
//     @@ -12,7 +12,7 @@
//                  "Params": [
//                      {
//                          "Embedded": false,
//     -                    "Name": ".arg_0",
//     +                    "Name": "cur",
//                          "PkgPath": "",
//                          "Tag": "",
//                          "Type": {
//     @@ -22,6 +22,10 @@
//                      }
//                  ],
//                  "Results": null
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "ObjectID": "05afb073ed788a666b933765c9bde3fa80fa7991:5"
//              }
//          }
//      }
// u[05afb073ed788a666b933765c9bde3fa80fa7991:5](7)=
//     @@ -8,10 +8,10 @@
//          "NativePkg": "",
//          "ObjectInfo": {
//              "ID": "05afb073ed788a666b933765c9bde3fa80fa7991:5",
//     +        "IsEscaped": true,
//              "LastObjectSize": "522",
//     -        "ModTime": "0",
//     -        "OwnerID": "05afb073ed788a666b933765c9bde3fa80fa7991:2",
//     -        "RefCount": "1"
//     +        "ModTime": "8",
//     +        "RefCount": "2"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// u[05afb073ed788a666b933765c9bde3fa80fa7991:4](111)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "05afb073ed788a666b933765c9bde3fa80fa7991:4",
//              "LastObjectSize": "207",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "05afb073ed788a666b933765c9bde3fa80fa7991:2",
//              "RefCount": "1"
//          },
//     @@ -11,6 +11,11 @@
//                  "@type": "/gno.FuncType",
//                  "Params": null,
//                  "Results": null
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "f2fba5f7ce6527a91f25c198ef772d40db5f04a9",
//     +            "ObjectID": "05afb073ed788a666b933765c9bde3fa80fa7991:9"
//              }
//          }
//      }
// u[05afb073ed788a666b933765c9bde3fa80fa7991:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "05afb073ed788a666b933765c9bde3fa80fa7991:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1010",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5ccb7c7915454745b491eac2e3292bca62c797db",
//     +                "Hash": "6d59834a85246828acf1c4ae78dac4c8dce81572",
//                      "ObjectID": "05afb073ed788a666b933765c9bde3fa80fa7991:3"
//                  }
//              },
//     @@ -44,7 +44,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "6f0999093f0621750e5b2633a62def1751a2d9de",
//     +                "Hash": "f090058bf303548dfbeeac7130b5e7d9fe29f0cd",
//                      "ObjectID": "05afb073ed788a666b933765c9bde3fa80fa7991:4"
//                  }
//              },
//     @@ -67,7 +67,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "afbe705201140f59ca7a9ac74593ad62b9ffe281",
//                      "ObjectID": "05afb073ed788a666b933765c9bde3fa80fa7991:5"
//                  }
//              },
