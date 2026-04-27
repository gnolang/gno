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
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "a", Value: 1})
}

func main(cur realm) {
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "c", Value: 3})
	println(crossrealm_f.Len())
}

// Output:
// 2

// Realm oplog legend (object IDs are from this run; package = crossrealm_f):
//   :2  PackageValue (escaped, RefCount=2)
//   :3  PackageBlock — holds the `entries` slice header
//   :9  OLD backing array (cap 1) from Tx1 — DELETED in Tx2 (see d[…:9])
//   :10 OLD HeapItemValue boxing *Entry "a" from Tx1 — its OwnerID still
//       points at :9. The u[…:10] update is markDirtyAncestors walking
//       through it; without GetObjectSafe, that walk would call
//       GetObject(:9) and panic because :9 is the d[…] line below.
//   :12 NEW backing array (cap 2) created by main's append
//   :13 NEW HeapItemValue boxing *Entry "c"
//   :14 NEW Entry{Key:"c", Value:3}
//
// The thing to notice is that d[…:9] (deleted) sits in the same
// finalization as u[…:10] (updated), and :10's OwnerID = :9. That
// adjacency is the bug-trigger: a dirty walk that follows :10 → :9.

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm_f"]
// c[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:14](279)={
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
//         "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:14",
//         "LastObjectSize": "279",
//         "ModTime": "0",
//         "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:13",
//         "RefCount": "1"
//     }
// }
// c[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:13](357)={
//     "ObjectInfo": {
//         "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:13",
//         "LastObjectSize": "357",
//         "ModTime": "0",
//         "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:12",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/tests/vm/crossrealm_f.Entry"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "79f3fd3402df0e9f2a0eb775c279bc92acfb6aa6",
//             "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:14"
//         }
//     }
// }
// c[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:12](620)={
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
//                     "Hash": "19af7863aa44e3f7c068b64ca6326094759a0c8b",
//                     "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:10"
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
//                     "Hash": "608f872f0da28cc47a80ba042f51ab914bd40a88",
//                     "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:13"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:12",
//         "LastObjectSize": "620",
//         "ModTime": "0",
//         "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:3",
//         "RefCount": "1"
//     }
// }
// u[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:3](1)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:3",
//              "LastObjectSize": "427",
//     -        "ModTime": "8",
//     +        "ModTime": "11",
//              "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:2",
//              "RefCount": "1"
//          },
//     @@ -22,11 +22,11 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5e471edfc36b38874518b170a59210e7b82d4610",
//     -                "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:9"
//     +                "Hash": "8385d2de0f0d107e369bc2558fb3eb91b66bd099",
//     +                "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:12"
//                  },
//     -            "Length": "1",
//     -            "Maxcap": "1",
//     +            "Length": "2",
//     +            "Maxcap": "2",
//                  "Offset": "0"
//              }
//          }
// u[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:10](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:10",
//              "LastObjectSize": "356",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:9",
//              "RefCount": "1"
//          },
// u[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1488",
//     -        "ModTime": "11",
//     +        "ModTime": "14",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "fc674e3e9ba4c4f86aac899d5494ad92a79cb645",
//     +                "Hash": "7ad436c16d6ec9c2f0e317f55d7b29debf19c977",
//                      "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:3"
//                  }
//              },
// d[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:9](-397)
// finalizerealm["gno.land/r/test/impl"]
