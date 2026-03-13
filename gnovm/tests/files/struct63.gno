// PKGPATH: gno.land/r/main
package main

type A struct {
	b B
}

type B struct {
	c C
}

type C struct {
	d D
}

type D struct {
	e int
}

var x = A{b: B{c: C{d: D{e: 1}}}}

func main() {
	x.b.c = C{d: D{e: 2}}
}

// Realm:
// finalizerealm["gno.land/r/main"]
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:11](216)={
//     "Fields": [
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:11",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:10",
//         "RefCount": "1"
//     }
// }
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:10](333)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/main.D"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "9ecaaba55c8400d0a9b7f90804afeebeccd80a79",
//                 "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:11"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:10",
//         "LastObjectSize": "333",
//         "ModTime": "0",
//         "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:5",
//         "RefCount": "1"
//     }
// }
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:5](6)=
//     @@ -7,15 +7,15 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "0838a9b6bc9a600cf3efff9dd11c009eda1a6cc2",
//     -                "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:6"
//     +                "Hash": "6a3c6de5e326bc3f5905a4223efd057103f1eec0",
//     +                "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:10"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:5",
//              "LastObjectSize": "331",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "RefCount": "1"
//          }
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4](5)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "aa2eddc3fe87537ded124ddb812d4dff7a05e912",
//     +                "Hash": "8fe8fdb193d18c74a2c80ecf0476fc8d901ffc40",
//                      "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:5"
//                  }
//              }
//     @@ -15,7 +15,7 @@
//          "ObjectInfo": {
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "LastObjectSize": "331",
//     -        "ModTime": "0",
//     +        "ModTime": "11",
//              "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "RefCount": "1"
//          }
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "LastObjectSize": "333",
//     -        "ModTime": "0",
//     +        "ModTime": "11",
//              "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "7de0b61008ea7b8caf1fed1f2fb08095c067b373",
//     +            "Hash": "9143975c6199c0f361c79ea8bffa57d3c916c784",
//                  "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4"
//              }
//          }
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1180",
//     -        "ModTime": "0",
//     +        "ModTime": "11",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -210,7 +210,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "ca2914f3d06d7976ed7cb34d3c08c9868c14848c",
//     +                "Hash": "87762033529689f6766c3b9f8e554f2a9460c7c5",
//                      "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3"
//                  }
//              },
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:6](-331)
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:7](-214)
