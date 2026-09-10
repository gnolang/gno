// PKGPATH: gno.land/r/test
package test

// ArrayValue.Copy shares the element values when the element type is an
// interface, rather than deep-copying them. When the element is created and
// the array copied within the same transaction, the element is therefore
// persisted once as a shared escaped object (RefCount 2, no OwnerID) instead
// of twice as two independently-owned objects. Escape is one-way, so pin the
// resulting realm layout.

type S struct{ F int }

var (
	arr  [1]any
	arr2 [1]any
)

func main(cur realm) {
	arr[0] = S{1} // fresh, unreal *StructValue
	arr2 = arr    // shares it: the element is not deep-copied
	println("vals:", arr[0].(S).F, arr2[0].(S).F)
}

// Output:
// vals: 1 1

// Realm:
// finalizerealm["gno.land/r/test"]
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:9](216)={
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
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "IsEscaped": true,
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "RefCount": "2"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:10](331)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test.S"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "LastObjectSize": "331",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "RefCount": "1"
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](148)=
//     @@ -1,12 +1,21 @@
//      {
//          "Data": null,
//          "List": [
//     -        {}
//     +        {
//     +            "T": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/test.S"
//     +            },
//     +            "V": {
//     +                "@type": "/gno.RefValue",
//     +                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9"
//     +            }
//     +        }
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "187",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:5](6)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5",
//              "LastObjectSize": "340",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -20,8 +20,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "95ae3fcadc857e60e745681015224a980bf48d4d",
//     -            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6"
//     +            "Hash": "fbf404ca893f3177d30fa41bb790f7eeff265cf5",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "340",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -20,7 +20,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "e8ff7ebef7a4fae35d190d478db945118eab9eca",
//     +            "Hash": "68172d6092c756d445fa84945d8177144e2e7cb5",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "762",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "97bdd03f97cb103ecacee3e9d1a366217c00eacf",
//     +                "Hash": "57e394e8d16bc64270451621e1faae2cfc0990e3",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
//     @@ -56,7 +56,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "39ee2ee9d799064ff1e52e466cd7b426d52d0c30",
//     +                "Hash": "9601a2647c9843967adcdcd9f656ec939c9fd0e6",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                  }
//              },
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:6](-187)
