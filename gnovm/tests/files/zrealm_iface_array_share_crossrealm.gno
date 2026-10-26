// PKGPATH: gno.land/r/crossrealm
package crossrealm

// Companion to zrealm_iface_array_share.gno, with the array handed to another
// realm by value. The receiving realm's stored copy shares the element rather
// than owning a copy of it, so the element escapes (RefCount 2, no OwnerID)
// and crossrealm_b's array holds a RefValue to it.

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
//                 "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8"
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
//     +            "Hash": "485b8fefe97444680199cdd60665c09caa745691",
//     +            "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:21"
//     +        }
//     +    }
//      }
// u[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8](7)=
//     @@ -10,9 +10,9 @@
//          ],
//          "ObjectInfo": {
//              "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8",
//     +        "IsEscaped": true,
//              "LastObjectSize": "214",
//     -        "ModTime": "0",
//     -        "OwnerID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:4",
//     -        "RefCount": "1"
//     +        "ModTime": "21",
//     +        "RefCount": "2"
//          }
//      }
// u[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:4](0)=
//     @@ -8,7 +8,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "345298e3891a9693fe4c2179e6f21365c4c11eec",
//                      "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8"
//                  }
//              }
//     @@ -16,7 +15,7 @@
//          "ObjectInfo": {
//              "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:4",
//              "LastObjectSize": "341",
//     -        "ModTime": "7",
//     +        "ModTime": "21",
//              "OwnerID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3",
//              "RefCount": "1"
//          }
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
//     +                "Hash": "c5f0406b1b5bf883986a07247ec131c3ddd7cf9d",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:13"
//                  }
//              },
// u[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3",
//              "LastObjectSize": "345",
//     -        "ModTime": "8",
//     +        "ModTime": "21",
//              "OwnerID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:2",
//              "RefCount": "1"
//          },
//     @@ -20,7 +20,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "226b4c101f9ee642fb131f559ee644aa8f2f542d",
//     +            "Hash": "b630fefb1b7500baab64345cdb4ef0f33a4ae398",
//                  "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:4"
//              }
//          }
// u[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:2",
//              "IsEscaped": true,
//              "LastObjectSize": "775",
//     -        "ModTime": "8",
//     +        "ModTime": "21",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "4f65431b2d42a1b462e1da3713826235677d67c5",
//     +                "Hash": "cc2aa87f5c5520112c0d5b3b3de52bf371093f0d",
//                      "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3"
//                  }
//              },
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/crossrealm"]
