// PKGPATH: gno.land/r/test/impl
package impl

// This test exercises getOwner returning nil for a deleted owner object.
//
// During init(), two entries are added to realm D's slice via append().
// Each append may grow the backing array, causing the old ArrayValue to
// be saved then deleted within the same finalization. However, a
// HeapItemValue created during the same init still has its OwnerID
// pointing to the now-deleted ArrayValue.
//
// When main() runs and adds another entry, markDirtyAncestors walks up
// from dirty objects and reaches that HeapItemValue. Its OwnerID points
// to the deleted ArrayValue, which no longer exists in the store.
//
// getOwner must use GetObjectSafe and return nil to stop the walk,
// rather than panic with GetObject.
//
// Without this fix: panic: unexpected object with id ...:9
// With this fix:    walk stops at nil, test passes.

import (
	"gno.land/r/tests/vm/crossrealm_f"
)

func init() {
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "a", Value: 1})
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "b", Value: 2})
}

func main(cur realm) {
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "c", Value: 3})
	println(crossrealm_f.Len())
}

// Output:
// 3

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm_f"]
// c[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:17](279)={
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
//         "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:17",
//         "LastObjectSize": "279",
//         "ModTime": "0",
//         "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:16",
//         "RefCount": "1"
//     }
// }
// c[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:16](357)={
//     "ObjectInfo": {
//         "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:16",
//         "LastObjectSize": "357",
//         "ModTime": "0",
//         "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:15",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/tests/vm/crossrealm_f.Entry"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "6cd34388ce3503828f99337381be942334f6c962",
//             "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:17"
//         }
//     }
// }
// c[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:15](842)={
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
//                     "Hash": "3b14c473fb0587ecc0be1525fe6b482546224d30",
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
//                     "Hash": "98ffa6505e55b6d515892efaa7a11899873278a0",
//                     "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:13"
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
//                     "Hash": "b5fcd45cf39ae7976150875e3568f45d85f0d646",
//                     "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:16"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:15",
//         "LastObjectSize": "842",
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
//     -        "ModTime": "11",
//     +        "ModTime": "14",
//              "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:2",
//              "RefCount": "1"
//          },
//     @@ -22,11 +22,11 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "03e7ccddbdc16db4d066d7719dc1c714025f78ba",
//     -                "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:12"
//     +                "Hash": "e82b6555ac84fc1414cd51beade803d87813c282",
//     +                "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:15"
//                  },
//     -            "Length": "2",
//     -            "Maxcap": "2",
//     +            "Length": "3",
//     +            "Maxcap": "3",
//                  "Offset": "0"
//              }
//          }
// u[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:10](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:10",
//              "LastObjectSize": "361",
//     -        "ModTime": "12",
//     +        "ModTime": "15",
//              "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:9",
//              "RefCount": "1"
//          },
// u[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:13](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:13",
//              "LastObjectSize": "357",
//     -        "ModTime": "0",
//     +        "ModTime": "15",
//              "OwnerID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:12",
//              "RefCount": "1"
//          },
// u[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1488",
//     -        "ModTime": "14",
//     +        "ModTime": "17",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "7134f697f7f0ebccae35552567d96eee9bc43ac9",
//     +                "Hash": "c6a51f41a307a9dba3ad3b232c005e62f601e0d6",
//                      "ObjectID": "0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:3"
//                  }
//              },
// d[0f5bbfc9234b76166575c9723efa3e0a3b15f9f8:12](-620)
// finalizerealm["gno.land/r/test/impl"]
