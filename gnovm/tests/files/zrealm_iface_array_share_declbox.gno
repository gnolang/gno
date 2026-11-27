// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Box is declared in this realm, so a copy of it made by crossrealm_b is
// stamped with this realm, not crossrealm_b. The element must still be
// deep-copied: crossrealm_b's array owns its own element, and this realm's
// element is left untouched (not escaped), as in
// zrealm_iface_array_share_crossrealm.gno.

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type Box [1]any

var arr Box

func init() {
	arr[0] = [2]int{1, 2}
}

func main(cur realm) {
	println("pre:", arr[0].([2]int)[0])
	crossrealm_b.SetObject(cross(cur), arr)
	got := crossrealm_b.GetObject().(Box)
	println("here:", arr[0].([2]int)[0], "there:", got[0].([2]int)[0])
}

// Output:
// pre: 1
// here: 1 there: 1

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// c[06279d1e03ecb38b84822f9caf584bd16913470d:21](252)={
//     "Data": null,
//     "List": [
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:21",
//         "LastObjectSize": "252",
//         "ModTime": "0",
//         "OwnerID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:9",
//         "RefCount": "1"
//     }
// }
// c[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:9](343)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.ArrayType",
//                 "Elt": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "32"
//                 },
//                 "Len": "2",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "ca579bbf0b63f48f3e3da693b9a576bedc51c7c4",
//                 "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:21"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:9",
//         "LastObjectSize": "343",
//         "ModTime": "0",
//         "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:13",
//         "RefCount": "1"
//     }
// }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:13](156)=
//     @@ -2,9 +2,19 @@
//          "ObjectInfo": {
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:13",
//              "LastObjectSize": "191",
//     -        "ModTime": "0",
//     +        "ModTime": "20",
//              "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.RefType",
//     +            "ID": "gno.land/r/crossrealm.Box"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "d7acbe129fb196062a7a08162f7a76fcc785a957",
//     +            "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:9"
//     +        }
//     +    }
//      }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "IsEscaped": true,
//              "LastObjectSize": "2495",
//     -        "ModTime": "0",
//     +        "ModTime": "21",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -120,7 +120,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1a1bad0b9e9c1282896dc99ce4e5ba6255cc3444",
//     +                "Hash": "0787850de951eea6240957a2b54bdea51ca25668",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:13"
//                  }
//              },
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/crossrealm"]
