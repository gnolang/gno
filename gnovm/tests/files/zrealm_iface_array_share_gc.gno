// PKGPATH: gno.land/r/test
// MAXALLOC: 4000000
package test

// A realm copies a stored array of interfaces, clears the source slot, then
// allocates enough garbage to trigger a GC before storing the copy. A shared
// stored element would be evicted from the cache by the GC and reloaded as a
// second in-memory object, which finalize then saves as escaped while also
// deleting it, leaving arr2 unreadable. The copy must own its own element.

type S struct{ F int }

var arr, arr2 [1]any

func init() {
	arr[0] = S{1}
}

func main(cur realm) {
	b := arr
	arr[0] = nil
	for i := 0; i < 40; i++ {
		_ = make([]byte, 200000) // garbage: pushes the allocator into a GC
	}
	arr2 = b
	println("ok")
}

// Output:
// ok

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
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](-144)=
//     @@ -1,22 +1,12 @@
//      {
//          "Data": null,
//          "List": [
//     -        {
//     -            "T": {
//     -                "@type": "/gno.RefType",
//     -                "ID": "gno.land/r/test.S"
//     -            },
//     -            "V": {
//     -                "@type": "/gno.RefValue",
//     -                "Hash": "3812e0e5a173468cbb34194e2bb8fe52d2b2d4cc",
//     -                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//     -            }
//     -        }
//     +        {}
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "336",
//     -        "ModTime": "9",
//     +        "ModTime": "10",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
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
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "345",
//     -        "ModTime": "10",
//     +        "ModTime": "12",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -20,7 +20,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "f381efd266fdd1421be26a0d236b366d08e27ae0",
//     +            "Hash": "67e24ebe6c688ceadd5e439839a249d8bbd04567",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//              }
//          }
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
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "85014db4921dcb092be79497dbf2c0ca878bc1d9",
//     +                "Hash": "5a3a53cee2d03c0b9e698a10e611a846e1da9b86",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
//     @@ -56,7 +56,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "39ee2ee9d799064ff1e52e466cd7b426d52d0c30",
//     +                "Hash": "ed64f496bc56ad27aee8568b6ccf5b110b289fe1",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                  }
//              },
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:10](-215)
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:6](-187)
