// PKGPATH: gno.land/r/map_key_alias
package map_key_alias

// A map key must be COPIED at insert, per Go semantics: mutating the source
// variable afterwards must not change the stored key. That only holds if the
// key's lazily-loaded children are resolved BEFORE the copy — otherwise
// Copy() shallow-copies an unresolved RefValue child and the later fill points
// the copy at the shared store object, aliasing the caller's graph.
//
// Requires a NESTED object-bearing field. A key whose fields are all
// primitives has no RefValue children and cannot reproduce this.

type inner struct {
	a int
}

type key struct {
	i inner
}

var (
	m map[key]int
	k key
)

func init() {
	m = map[key]int{}
	k = key{i: inner{a: 1}}
	m[k] = 10
}

func main(cur realm) {
	m[k] = 20 // re-assign the existing key, in a later tx
	k.i.a = 5 // mutate the source variable afterwards
	for sk, v := range m {
		println("stored key:", sk.i.a, "value:", v)
	}
}

// Output:
// stored key: 1 value: 20

// Realm:
// finalizerealm["gno.land/r/map_key_alias"]
// c[057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:16](216)={
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
//         "ID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:16",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:15",
//         "RefCount": "1"
//     }
// }
// c[057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:15](347)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/map_key_alias.inner"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "40b48951ac9984bf4e6692b96ae9b108cc3feab9",
//                 "ObjectID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:16"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:15",
//         "LastObjectSize": "347",
//         "ModTime": "0",
//         "OwnerID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:10",
//         "RefCount": "1"
//     }
// }
// u[057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:10](5)=
//     @@ -9,12 +9,12 @@
//                          },
//                          "V": {
//                              "@type": "/gno.RefValue",
//     -                        "Hash": "81247205b765a08ee804fde5f829c971061d4c80",
//     -                        "ObjectID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:11"
//     +                        "Hash": "57c8b3db36240d62b2e78066367504906af925e3",
//     +                        "ObjectID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:15"
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
//              "ID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:10",
//              "LastObjectSize": "385",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:3",
//              "RefCount": "1"
//          }
// u[057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:14](5)=
//     @@ -1,7 +1,7 @@
//      {
//          "Fields": [
//              {
//     -            "N": "AQAAAAAAAAA=",
//     +            "N": "BQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -11,7 +11,7 @@
//          "ObjectInfo": {
//              "ID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:14",
//              "LastObjectSize": "216",
//     -        "ModTime": "0",
//     +        "ModTime": "14",
//              "OwnerID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:13",
//              "RefCount": "1"
//          }
// u[057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:3",
//              "LastObjectSize": "394",
//     -        "ModTime": "9",
//     +        "ModTime": "16",
//              "OwnerID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:2",
//              "RefCount": "1"
//          },
//     @@ -20,7 +20,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "46447572d6b05567144a567cc7bdccae4fca6358",
//     +            "Hash": "be1f2e44805a89f6d89be4a75c7cc0be47c84ee6",
//                  "ObjectID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:10"
//              }
//          }
// u[057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1021",
//     -        "ModTime": "14",
//     +        "ModTime": "16",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -58,7 +58,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "302cb9fe0902353a488c0ef47fa42b402e3f34ad",
//     +                "Hash": "fee61e4c4b778baff0f4cb8a6f8b51014c919177",
//                      "ObjectID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:3"
//                  }
//              },
//     @@ -68,7 +68,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "90262307aec6ade441d95776d717f1992c669e88",
//                      "ObjectID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:4"
//                  }
//              },
// u[057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:13](5)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "7dfdb6333c241e2c4844d6fc17362af74edfe55f",
//     +                "Hash": "2fba846c21438b309173b1b8cc48b8b7c59e2ff9",
//                      "ObjectID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:14"
//                  }
//              }
//     @@ -15,7 +15,7 @@
//          "ObjectInfo": {
//              "ID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:13",
//              "LastObjectSize": "346",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:4",
//              "RefCount": "1"
//          }
// u[057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:4](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:4",
//              "LastObjectSize": "350",
//     -        "ModTime": "9",
//     +        "ModTime": "16",
//              "OwnerID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:2",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "35b61dfc31a85091f206bdc16badc047233ef78b",
//     +            "Hash": "903d71c87a418cc8c90b1bce11789556cb334312",
//                  "ObjectID": "057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:13"
//              }
//          }
// d[057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:11](-347)
// d[057fc2feeab7ed73fc65eb9207f2a7ee3a3a5798:12](-216)
