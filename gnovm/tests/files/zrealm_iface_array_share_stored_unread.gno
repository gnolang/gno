// PKGPATH: gno.land/r/test
package test

// Companion to zrealm_iface_array_share_stored.gno: the same copy, without
// reading the element first, so ArrayValue.Copy sees the unloaded RefValue.
// The realm layout must match that test's: the element is loaded and
// deep-copied, and arr2 owns its own object.

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
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:12](216)={
//     "Fields": [
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "RefCount": "1"
//     }
// }
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
//                 "Hash": "b15f4f1988caca393d4c3ca0d8a01e0fae0c6438",
//                 "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12"
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
//     +            "Hash": "9de22d8dbc328742ea7920a8d6ff0cd4ae1e5550",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "902",
//     -        "ModTime": "10",
//     +        "ModTime": "12",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -56,7 +56,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "39ee2ee9d799064ff1e52e466cd7b426d52d0c30",
//     +                "Hash": "ed64f496bc56ad27aee8568b6ccf5b110b289fe1",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                  }
//              },
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:6](-187)
