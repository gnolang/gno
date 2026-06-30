// PKGPATH: gno.land/r/test
package test

// GVM-01 regression: an append-into-capacity write to a List-backed []byte
// realm variable must mark the backing ArrayValue dirty via DidUpdate, so the
// mutation persists. Previously the copyDataToList sub-branch in uverse.go
// wrote raw bytes into the array's .List without calling DidUpdate, so the
// array was absent from the realm commit diff and the write was dropped.
//
// A composite-literal []byte is List-backed (unlike []byte("..."), which is
// Data-backed); append(b[:1], "Z"...) takes the List-dst + Data-src branch.
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
//     @@ -8,6 +8,7 @@
//                  }
//              },
//              {
//     +            "N": "WgAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "4096"
//     @@ -29,7 +30,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "331",
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
//     -                "Hash": "27feff93f5aa5d0a0fb9fdb689ef94e977fa794a",
//     +                "Hash": "fa62c43e66bf8a51800fb8ed91ac0aaaae8ac805",
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
//     -                "Hash": "dacaff18ab949ca294c0ef65d8fa6211de74f30f",
//     +                "Hash": "7a1c191d2d80ba27de71b8455bf65d4661b9e743",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
