// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm"
	"gno.land/r/tests/vm/crossrealm_b"
)

func main(cur realm) {
	f := crossrealm_b.Fooer
	crossrealm.SetFooer(cross, f)
	crossrealm.CallFooerFooCross(cross)
	f.SetS("B")
	crossrealm.CallFooerFooCross(cross)
	print(".")
}

// Output:
// hello A cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// hello B cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:7](214)=
//     @@ -2,9 +2,27 @@
//          "ObjectInfo": {
//              "ID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:7",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "43",
//              "OwnerID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "N": "UmVhRG9OTFk=",
//     +        "T": {
//     +            "@type": "/gno.PointerType",
//     +            "Elt": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/tests/vm/crossrealm_b.fooer"
//     +            }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:4"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//     +        }
//     +    }
//      }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:4](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//     +        "IsEscaped": true,
//              "LastObjectSize": "354",
//     -        "ModTime": "0",
//     -        "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:3",
//     -        "RefCount": "1"
//     +        "ModTime": "43",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[06279d1e03ecb38b84822f9caf584bd16913470d:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:3",
//              "LastObjectSize": "399",
//     -        "ModTime": "0",
//     +        "ModTime": "43",
//              "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,6 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "2cf29cfddfad3bf584a0b0607697df3d23ecf04c",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:4"
//                  },
//                  "Index": "0",
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:2](0)=
//     @@ -105,7 +105,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "d49b3fb9d433be706f6d81b5bbf3855107b513ca",
//     +                "Hash": "c924dfae8c78109d347797a67a94ddffbd1f1295",
//                      "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:7"
//                  }
//              },
// u[06279d1e03ecb38b84822f9caf584bd16913470d:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "IsEscaped": true,
//              "LastObjectSize": "2103",
//     -        "ModTime": "0",
//     +        "ModTime": "43",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e6301dc45d62d2b7659c3f5fff00f83b0a69f7d7",
//     +                "Hash": "58b6ef7fbd90cca4b6157fb97f9017b4bb55d0d2",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:3"
//                  }
//              },
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[06279d1e03ecb38b84822f9caf584bd16913470d:5](5)=
//     @@ -7,14 +7,14 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "A"
//     +                "value": "B"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:5",
//              "LastObjectSize": "239",
//     -        "ModTime": "0",
//     +        "ModTime": "18",
//              "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "RefCount": "1"
//          }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -3,7 +3,7 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "IsEscaped": true,
//              "LastObjectSize": "361",
//     -        "ModTime": "43",
//     +        "ModTime": "18",
//              "RefCount": "2"
//          },
//          "Value": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "534961c4c33e00090c5b38e73cdddaa534514e00",
//     +            "Hash": "cf8abd9e6c10a053c7eccb921d67b74c5505729c",
//                  "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]
