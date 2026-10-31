// PKGPATH: gno.land/r/test
package test

// Companion to zrealm_iface_array_share_stored.gno: the same copy, without
// reading the element first, so ArrayValue.Copy sees the unloaded RefValue.
// The realm layout must match that test's: the element is shared and escapes.

type S struct{ F int }

var (
	arr  [1]any
	arr2 [1]any
)

func init() {
	arr[0] = S{1}
}

func main(cur realm) {
	arr2 = arr
	println("vals:", arr[0].(S).F, arr2[0].(S).F)
}

// Output:
// vals: 1 1

// Realm:
// finalizerealm["gno.land/r/test"]
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:11](332)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test.S"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "LastObjectSize": "332",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "RefCount": "1"
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:5](6)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5",
//              "LastObjectSize": "340",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -20,8 +20,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "95ae3fcadc857e60e745681015224a980bf48d4d",
//     -            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6"
//     +            "Hash": "c1d06f649898b5409b90c35ca7f012f973bfe6f6",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:10](7)=
//     @@ -10,9 +10,9 @@
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//     +        "IsEscaped": true,
//              "LastObjectSize": "215",
//     -        "ModTime": "0",
//     -        "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//     -        "RefCount": "1"
//     +        "ModTime": "11",
//     +        "RefCount": "2"
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -8,7 +8,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "3812e0e5a173468cbb34194e2bb8fe52d2b2d4cc",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//                  }
//              }
//     @@ -16,7 +15,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "336",
//     -        "ModTime": "9",
//     +        "ModTime": "11",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "902",
//     -        "ModTime": "10",
//     +        "ModTime": "11",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "85014db4921dcb092be79497dbf2c0ca878bc1d9",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
//     @@ -56,7 +55,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "39ee2ee9d799064ff1e52e466cd7b426d52d0c30",
//     +                "Hash": "a818a72bb2df19be19bafa1e6eed5dd16694d34e",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                  }
//              },
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "345",
//     -        "ModTime": "10",
//     +        "ModTime": "11",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -20,7 +20,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "f381efd266fdd1421be26a0d236b366d08e27ae0",
//     +            "Hash": "9252a6d0d5b2faeeb045171b729082b80f54479e",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//              }
//          }
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:6](-187)
