// PKGPATH: gno.land/r/test
package test

// GVM-01 regression: an append-into-capacity write to a []byte realm variable
// must mark the backing ArrayValue dirty via DidUpdate, so the mutation
// persists. Previously the raw-byte sub-branches in uverse.go wrote into the
// array without calling DidUpdate, so the array was absent from the realm
// commit diff and the write was dropped.
//
// Every []byte is Data-backed now (composite literals included), so
// append(b[:1], "Z"...) takes the Data-dst + Data-src branch. The List-dst
// branch this test originally covered is no longer reachable from Gno source.
var b = []byte{0, 0, 0, 0}

func main(cur realm) {
	_ = append(b[:1], "Z"...) // mutates b[1] within capacity (Z == 90)
	println(b[1])
}

// Output:
// 90

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](5)=
//     @@ -1,10 +1,10 @@
//      {
//     -    "Data": "AAAAAA==",
//     +    "Data": "AFoAAA==",
//          "List": null,
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "181",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "369",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -19,7 +19,7 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "7d99e547c6714ec970a8d1d7ef347b5be0f375e6",
//     +                "Hash": "600e8853f7b498737742b46001a2f767638ee2ae",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                  },
//                  "Length": "4",
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
//     -                "Hash": "1470b50e5e5f95e1473d33b59da47e4a71fc6cca",
//     +                "Hash": "a5580c7c34a508d8541176f52bb97ee86344e94a",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
