// PKGPATH: gno.land/r/test/impl
package impl

// This test exercises getOwner returning nil for a deleted owner object.
//
// Tx1 (init): one entry "a" is appended to realm F's slice. This persists:
//   - an ArrayValue (cap 1) holding the slice's backing storage
//   - a HeapItemValue boxing *Entry "a", whose OwnerID points to that array
//
// Tx2 (main): append("c") grows the backing array. The cap-1 array from
// Tx1 is deleted in Tx2's finalization and replaced by a new cap-N array.
// The HeapItemValue from Tx1 is untouched on disk and still carries
// OwnerID = (now-deleted cap-1 array).
//
// During Tx2's finalization, markDirtyAncestors walks up from dirty
// objects and reaches that HeapItemValue. Following its OwnerID with
// the unsafe GetObject would panic — the object is gone. The fix is
// to call GetObjectSafe and treat a nil result as "stop the walk."
//
// Without this fix: panic: unexpected object with id ...:N
// With this fix:    walk stops at nil, test passes.

import (
	"gno.land/r/tests/vm/crossrealm_f"
)

func init() {
	crossrealm_f.Add(cross, crossrealm_f.NewEntry("a", 1))
}

func main(cur realm) {
	crossrealm_f.Add(cross, crossrealm_f.NewEntry("c", 3))
	println(crossrealm_f.Len())
}

// Output:
// 2



// Realm oplog legend (object IDs are from this run; package = crossrealm_f):
//   :2  PackageValue (escaped, RefCount=2)
//   :3  PackageBlock — holds the `entries` slice header
//   :10 OLD backing array (cap 1) from Tx1 — DELETED in Tx2 (see d[…:10])
//   :11 OLD HeapItemValue boxing *Entry "a" from Tx1 — its OwnerID still
//       points at :10. The u[…:11] update is markDirtyAncestors walking
//       through it; without GetObjectSafe, that walk would call
//       GetObject(:10) and panic because :10 is the d[…] line below.
//   :13 NEW backing array (cap 2) created by main's append
//   :14 NEW HeapItemValue boxing *Entry "c"
//   :15 NEW Entry{Key:"c", Value:3}
//
// The thing to notice is that d[…:10] (deleted) sits in the same
// finalization as u[…:11] (updated), and :11's OwnerID = :10. That
// adjacency is the bug-trigger: a dirty walk that follows :11 → :10.

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm_f"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_f"]
// c[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:15](279)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "c"
//             }
//         },
//         {
//             "N": "AwAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:15",
//         "LastObjectSize": "279",
//         "ModTime": "0",
//         "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:14",
//         "RefCount": "1"
//     }
// }
// c[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:14](357)={
//     "ObjectInfo": {
//         "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:14",
//         "LastObjectSize": "357",
//         "ModTime": "0",
//         "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:13",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/tests/vm/crossrealm_f.Entry"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "6a5bc2b54e5678a9df79f5cf4b7b347111643a7f",
//             "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:15"
//         }
//     }
// }
// c[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:13](620)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests/vm/crossrealm_f.Entry"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "d046411631e5bf6eabd95a26d5d5b2eff5e12d08",
//                     "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:11"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests/vm/crossrealm_f.Entry"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "d95c3f5a9d3e8ca1813f81cb7bdc72b9c1d0874f",
//                     "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:14"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:13",
//         "LastObjectSize": "620",
//         "ModTime": "0",
//         "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:3",
//         "RefCount": "1"
//     }
// }
// u[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:3",
//              "LastObjectSize": "428",
//     -        "ModTime": "9",
//     +        "ModTime": "12",
//              "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:2",
//              "RefCount": "1"
//          },
//     @@ -22,11 +22,11 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "76e9a606e58a861202c90ab2ba062183a805ca8e",
//     -                "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:10"
//     +                "Hash": "992254e1ad027075482276f31ef24dddd0ffba15",
//     +                "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:13"
//                  },
//     -            "Length": "1",
//     -            "Maxcap": "1",
//     +            "Length": "2",
//     +            "Maxcap": "2",
//                  "Offset": "0"
//              }
//          }
// u[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:11](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:11",
//              "LastObjectSize": "357",
//     -        "ModTime": "0",
//     +        "ModTime": "13",
//              "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:10",
//              "RefCount": "1"
//          },
// u[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1785",
//     -        "ModTime": "12",
//     +        "ModTime": "15",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "3001cc01d53a7123a9707e04db0b7e27c5a2d5be",
//     +                "Hash": "34ed0964e359d8798c8f599ad2bb7e3435d2791f",
//                      "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:3"
//                  }
//              },
// d[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:10](-398)
// finalizerealm["gno.land/r/tests/vm/crossrealm_f"]
// finalizerealm["gno.land/r/test/impl"]
