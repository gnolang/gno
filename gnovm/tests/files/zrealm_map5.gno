// PKGPATH: gno.land/r/test
package test

// Regression: deleting an entry whose KEY is a non-primitive (array/struct)
// object from a persisted realm map must reclaim the stored key object — it
// must be emitted as a deleted object (d[...]) in the finalize diff, not
// orphaned. Before the fix the stored [1]int key object leaked (delete
// detached the argument key, not the stored key).
var m map[[1]int]int

func init() {
	m = map[[1]int]int{[1]int{1}: 10, [1]int{2}: 20}
}

func main(cur realm) {
	delete(m, [1]int{1})
	println("len:", len(m))
}

// Output:
// len: 1

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:7](-202)=
//     @@ -1,31 +1,6 @@
//      {
//          "List": {
//              "List": [
//     -            {
//     -                "Key": {
//     -                    "T": {
//     -                        "@type": "/gno.ArrayType",
//     -                        "Elt": {
//     -                            "@type": "/gno.PrimitiveType",
//     -                            "value": "32"
//     -                        },
//     -                        "Len": "1",
//     -                        "Vrd": false
//     -                    },
//     -                    "V": {
//     -                        "@type": "/gno.RefValue",
//     -                        "Hash": "983a2dfd65f409988f2165b8b8e31a3f0147a0d7",
//     -                        "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8"
//     -                    }
//     -                },
//     -                "Value": {
//     -                    "N": "CgAAAAAAAAA=",
//     -                    "T": {
//     -                        "@type": "/gno.PrimitiveType",
//     -                        "value": "32"
//     -                    }
//     -                }
//     -            },
//                  {
//                      "Key": {
//                          "T": {
//     @@ -56,7 +31,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "590",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "393",
//     -        "ModTime": "6",
//     +        "ModTime": "9",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -25,7 +25,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "a669cab1a0b6e73131ef8ee77e2ed6bedcf482e3",
//     +            "Hash": "b8ca988c204aac98121d376916a114a3f0bee2cd",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e000ab9359964c6da1719766969d91e0ab8de8dd",
//     +                "Hash": "b3f9131b15661e4bfa46d47af69a74439c92d91c",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
// d[08ada09dee16d791fd406d629fe29bb0ed084a30:8](-213)
