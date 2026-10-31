// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Companion to zrealm_iface_array_share_crossrealm.gno, with an element
// allocated by this realm in the same transaction and handed to another realm
// by value. The element is stamped with this realm, not the copy's, so it is
// deep-copied rather than shared: crossrealm_b persists (and pays the storage
// deposit for) its own element, as an object of crossrealm_b.

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

func main(cur realm) {
	crossrealm_b.SetObject(cross(cur), [1]any{[2]int{1, 2}})
	got := crossrealm_b.GetObject().([1]any)
	println("there:", got[0].([2]int)[0])
}

// Output:
// there: 1

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// c[06279d1e03ecb38b84822f9caf584bd16913470d:22](253)={
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
//         "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:22",
//         "LastObjectSize": "253",
//         "ModTime": "0",
//         "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:21",
//         "RefCount": "1"
//     }
// }
// c[06279d1e03ecb38b84822f9caf584bd16913470d:21](344)={
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
//                 "Hash": "06a0f24747f250f215c5a21d90888f3374547fe3",
//                 "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:22"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:21",
//         "LastObjectSize": "344",
//         "ModTime": "0",
//         "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:13",
//         "RefCount": "1"
//     }
// }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:13](156)=
//     @@ -2,9 +2,26 @@
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
//     +            "@type": "/gno.ArrayType",
//     +            "Elt": {
//     +                "@type": "/gno.InterfaceType",
//     +                "Generic": "",
//     +                "Methods": null,
//     +                "PkgPath": ""
//     +            },
//     +            "Len": "1",
//     +            "Vrd": false
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "05952813b2f69d742ebc19a9749a8ee5774856f0",
//     +            "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:21"
//     +        }
//     +    }
//      }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "IsEscaped": true,
//              "LastObjectSize": "2495",
//     -        "ModTime": "0",
//     +        "ModTime": "22",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -120,7 +120,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1a1bad0b9e9c1282896dc99ce4e5ba6255cc3444",
//     +                "Hash": "542df1b8768fcb81d4aa0fac7a21e9619c80d419",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:13"
//                  }
//              },
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/crossrealm"]
