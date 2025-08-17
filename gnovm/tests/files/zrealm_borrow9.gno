// PKGPATH: gno.land/r/borrow
package borrow

import "gno.land/r/demo/tests/ext"

type NoopMutator struct{ num int }

func (nm *NoopMutator) Mutate() {
	println("==========Mutate, maybe malicous...")
}

var root interface{}

func main() {
	var nm = &NoopMutator{num: 1}
	root = nm
	ext.SetMutator(cross, nm) // attached to ext firstly.
}

// Realm:
// finalizerealm["gno.land/r/demo/tests/ext"]
// c[4f26ea8f8f08e71800f22b46c09652d382e95799:13](216)={
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
//         "ID": "4f26ea8f8f08e71800f22b46c09652d382e95799:13",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "4f26ea8f8f08e71800f22b46c09652d382e95799:12",
//         "RefCount": "1"
//     }
// }
// c[4f26ea8f8f08e71800f22b46c09652d382e95799:12](349)={
//     "ObjectInfo": {
//         "ID": "4f26ea8f8f08e71800f22b46c09652d382e95799:12",
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
//             "Hash": "ff5603030df8a3b74c19996fe7fa7717d1552cda",
//             "ObjectID": "4f26ea8f8f08e71800f22b46c09652d382e95799:13"
//         }
//     }
// }
// u[4f26ea8f8f08e71800f22b46c09652d382e95799:3](209)=
//     @@ -2,9 +2,27 @@
//          "ObjectInfo": {
//              "ID": "4f26ea8f8f08e71800f22b46c09652d382e95799:3",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "11",
//              "OwnerID": "4f26ea8f8f08e71800f22b46c09652d382e95799:2",
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
//     +                "ObjectID": "4f26ea8f8f08e71800f22b46c09652d382e95799:12"
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
//     +                "ObjectID": "4f26ea8f8f08e71800f22b46c09652d382e95799:12"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//     +        }
//     +    }
//      }
