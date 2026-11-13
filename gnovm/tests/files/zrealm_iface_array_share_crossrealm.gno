// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Companion to zrealm_iface_array_share.gno, with the stored array handed to
// another realm by value. The element is stamped with this realm, not the
// copy's, so it is deep-copied rather than shared: crossrealm_b's array owns
// its own element, and this realm's element is left untouched (not escaped).

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type S struct{ F int }

var arr [1]any

func init() {
	arr[0] = S{1}
}

func main(cur realm) {
	println("pre:", arr[0].(S).F)
	crossrealm_b.SetObject(cross(cur), arr)
	got := crossrealm_b.GetObject().([1]any)
	println("here:", arr[0].(S).F, "there:", got[0].(S).F)
}

// Output:
// pre: 1
// here: 1 there: 1

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// c[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:9](215)={
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
//         "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:9",
//         "LastObjectSize": "215",
//         "ModTime": "0",
//         "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:21",
//         "RefCount": "1"
//     }
// }
// c[06279d1e03ecb38b84822f9caf584bd16913470d:21](338)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/crossrealm.S"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "c7c1c84bb3209709659ba33ffc3d9a9b662c0fbd",
//                 "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:9"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:21",
//         "LastObjectSize": "338",
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
//     +            "Hash": "296d06984a97dfd4c3e97e6e89c5e29086904193",
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
//     +        "ModTime": "21",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -120,7 +120,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1a1bad0b9e9c1282896dc99ce4e5ba6255cc3444",
//     +                "Hash": "5d638d69bf82ae62397cf9858cd1fb8731677c79",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:13"
//                  }
//              },
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/crossrealm"]
