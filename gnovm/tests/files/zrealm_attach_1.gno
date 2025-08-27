// PKGPATH: gno.land/r/borrow
package borrow

import crossrealm "gno.land/r/demo/tests/crossrealm_d"

type NoopMutator struct{ num int }

func (nm *NoopMutator) Mutate() {
	println("==========Mutate, maybe malicous...")
}

var root interface{}

func main() {
	var nm = &NoopMutator{num: 1}
	crossrealm.SetMutator(cross, nm) // attached to ext firstly.
	root = nm                        // escape here...
}

// Realm:
// finalizerealm["gno.land/r/demo/tests/crossrealm_d"]
// c[0deeee8113af63ad50ebdabff9de7675ae4b2ea8:18](216)={
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
//         "ID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:18",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:17",
//         "RefCount": "1"
//     }
// }
// c[0deeee8113af63ad50ebdabff9de7675ae4b2ea8:17](347)={
//     "ObjectInfo": {
//         "ID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:17",
//         "LastObjectSize": "347",
//         "ModTime": "0",
//         "OwnerID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:3",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/borrow.NoopMutator"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "e5c70d6630b2a6e73c7ff308c5a475b51555a50a",
//             "ObjectID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:18"
//         }
//     }
// }
// u[0deeee8113af63ad50ebdabff9de7675ae4b2ea8:3](206)=
//     @@ -2,9 +2,27 @@
//          "ObjectInfo": {
//              "ID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:3",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "OwnerID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.PointerType",
//     +            "Elt": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/borrow.NoopMutator"
//     +            }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "29444b43ffb94bb8c0ce91994b7445071ff946b8",
//     +                "ObjectID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:17"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//     +        }
//     +    }
//      }
// finalizerealm["gno.land/r/borrow"]
// u[c459c7103cd5303d426ed1e6e642fffcffd55f7c:3](209)=
//     @@ -2,9 +2,27 @@
//          "ObjectInfo": {
//              "ID": "c459c7103cd5303d426ed1e6e642fffcffd55f7c:3",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "c459c7103cd5303d426ed1e6e642fffcffd55f7c:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.PointerType",
//     +            "Elt": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/borrow.NoopMutator"
//     +            }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Escaped": true,
//     +                "ObjectID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:17"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//     +        }
//     +    }
//      }
// u[0deeee8113af63ad50ebdabff9de7675ae4b2ea8:17](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:17",
//     +        "IsEscaped": true,
//              "LastObjectSize": "347",
//     -        "ModTime": "0",
//     -        "OwnerID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:3",
//     -        "RefCount": "1"
//     +        "ModTime": "5",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[0deeee8113af63ad50ebdabff9de7675ae4b2ea8:3](3)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:3",
//              "LastObjectSize": "396",
//     -        "ModTime": "16",
//     +        "ModTime": "5",
//              "OwnerID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "29444b43ffb94bb8c0ce91994b7445071ff946b8",
//     +                "Escaped": true,
//                      "ObjectID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:17"
//                  },
//                  "Index": "0",
