// PKGPATH: gno.land/r/map_key_dec_ref
package map_key_dec_ref

// Reassigning an EXISTING object-bearing key in a map that has round-tripped
// through the store must DecRef and delete the displaced stored key object.
// The vmap is built lazily, so the direct vmap read in GetPointerAtIndex needs
// an ensureVmap first; without it oldObject is silently nil, DidUpdate skips
// the DecRefCount branch, and the superseded key is orphaned with a stale
// RefCount — losing its storage-deposit refund as well.
//
// All three conditions must hold to trigger it: an object-bearing key (a
// struct, not a primitive), a map persisted across a tx boundary, and a first
// keyed operation after load that WRITES to an already-present key with no
// preceding read.

type key struct {
	a int
}

var m map[key]int

func init() {
	m = map[key]int{}
	m[key{1}] = 10
}

func main(cur realm) {
	m[key{1}] = 20 // reassign an existing key; no read first
	println(len(m))
	println(m[key{1}])
}

// Output:
// 1
// 20

// Realm:
// finalizerealm["gno.land/r/map_key_dec_ref"]
// c[01b8269e630c446eb86e2700d06299bb0c3f2a4e:9](214)={
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
//         "ID": "01b8269e630c446eb86e2700d06299bb0c3f2a4e:9",
//         "LastObjectSize": "214",
//         "ModTime": "0",
//         "OwnerID": "01b8269e630c446eb86e2700d06299bb0c3f2a4e:7",
//         "RefCount": "1"
//     }
// }
// u[01b8269e630c446eb86e2700d06299bb0c3f2a4e:7](5)=
//     @@ -9,12 +9,12 @@
//                          },
//                          "V": {
//                              "@type": "/gno.RefValue",
//     -                        "Hash": "b3197258a0e506a98399a46895f1000389b2db00",
//     -                        "ObjectID": "01b8269e630c446eb86e2700d06299bb0c3f2a4e:8"
//     +                        "Hash": "d0dd1734b7d55f682c0b2da0e253e47c6ad11c0d",
//     +                        "ObjectID": "01b8269e630c446eb86e2700d06299bb0c3f2a4e:9"
//                          }
//                      },
//                      "Value": {
//     -                    "N": "CgAAAAAAAAA=",
//     +                    "N": "FAAAAAAAAAA=",
//                          "T": {
//                              "@type": "/gno.PrimitiveType",
//                              "value": "32"
//     @@ -26,7 +26,7 @@
//          "ObjectInfo": {
//              "ID": "01b8269e630c446eb86e2700d06299bb0c3f2a4e:7",
//              "LastObjectSize": "385",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "01b8269e630c446eb86e2700d06299bb0c3f2a4e:3",
//              "RefCount": "1"
//          }
// u[01b8269e630c446eb86e2700d06299bb0c3f2a4e:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "01b8269e630c446eb86e2700d06299bb0c3f2a4e:3",
//              "LastObjectSize": "395",
//     -        "ModTime": "6",
//     +        "ModTime": "9",
//              "OwnerID": "01b8269e630c446eb86e2700d06299bb0c3f2a4e:2",
//              "RefCount": "1"
//          },
//     @@ -20,7 +20,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "fac03e894e69b16de276b3c779cc25f036c9771b",
//     +            "Hash": "814bba607bac3f4cee02681467c412a122d51716",
//                  "ObjectID": "01b8269e630c446eb86e2700d06299bb0c3f2a4e:7"
//              }
//          }
// u[01b8269e630c446eb86e2700d06299bb0c3f2a4e:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "01b8269e630c446eb86e2700d06299bb0c3f2a4e:2",
//              "IsEscaped": true,
//              "LastObjectSize": "787",
//     -        "ModTime": "8",
//     +        "ModTime": "9",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "d6bac97bec64be27ea0cad750facf57578207653",
//     +                "Hash": "0384f4b27539c4ee6589ad5149294cd482db5c5f",
//                      "ObjectID": "01b8269e630c446eb86e2700d06299bb0c3f2a4e:3"
//                  }
//              },
// d[01b8269e630c446eb86e2700d06299bb0c3f2a4e:8](-214)
