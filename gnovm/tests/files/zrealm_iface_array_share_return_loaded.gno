// PKGPATH: gno.land/r/test
package test

// Companion to zrealm_iface_array_share_return_unread.gno, with the element
// read before the copy, so ArrayValue.Copy sees a loaded, real *StructValue.
// The caller then stores the returned copy: it must own its own element
// rather than refer to the one the crossing return deleted.

type S struct{ F int }

var (
	arr  [1]any
	keep [1]any
)

func init() {
	arr[0] = S{1}
}

func Take(cur realm) [1]any {
	println("pre:", arr[0].(S).F) // loads the element
	b := arr
	arr[0] = nil
	return b
}

func main(cur realm) {
	b := Take(cross(cur))
	println("got:", b[0].(S).F)
	keep = b
}

// Output:
// pre: 1
// got: 1

// Realm:
// finalizerealm["gno.land/r/test"]
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
//     -                "Hash": "4ba9c7a015b4fe8dccf1a4504fc582746817b12a",
//     -                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11"
//     -            }
//     -        }
//     +        {}
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "336",
//     -        "ModTime": "10",
//     +        "ModTime": "11",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -20,7 +20,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "468cfb35911f4ccd9f00f4f0316dbace59c4fd48",
//     +            "Hash": "10628b2e62a8ed820a76d90fcecd68f8212f499c",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "49e0748cd2a2721a10554e8da4996190894bf24a",
//     +                "Hash": "ded75eda30081e329e6225387699b40c90280d2a",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:11](-215)
// finalizerealm["gno.land/r/test"]
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:13](216)={
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
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:12](332)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test.S"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "c06bc21919bf0ec29ac3dcc2036c4a5992ec22c3",
//                 "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
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
//     +        "ModTime": "11",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -20,8 +20,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "95ae3fcadc857e60e745681015224a980bf48d4d",
//     -            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6"
//     +            "Hash": "203925da1d5eeb5028852c3d1ab8fcc7458e5c2f",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1134",
//     -        "ModTime": "11",
//     +        "ModTime": "13",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -56,7 +56,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "39ee2ee9d799064ff1e52e466cd7b426d52d0c30",
//     +                "Hash": "317e0b6aaa7cd4c1ba774e9b7a718929aa4b9787",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                  }
//              },
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:6](-187)
