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
	root = nm
	crossrealm.SetMutator(cross, nm) // attached to crossrealm firstly.
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
// c[0deeee8113af63ad50ebdabff9de7675ae4b2ea8:17](349)={
//     "ObjectInfo": {
//         "ID": "0deeee8113af63ad50ebdabff9de7675ae4b2ea8:17",
//         "IsEscaped": true,
//         "LastObjectSize": "349",
//         "ModTime": "0",
//         "RefCount": "2"
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
// u[0deeee8113af63ad50ebdabff9de7675ae4b2ea8:3](209)=
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
//     +                "Escaped": true,
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
