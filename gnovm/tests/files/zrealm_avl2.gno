// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/nt/avl/v0"
)

var tree avl.Tree

func init() {
	tree.Set("key1", "value1")
}

func main(cur realm) {
	var updated bool
	updated = tree.Set("key2", "value2")
	// println(tree, updated)
	println(updated, tree.Size())
}

// Output:
// false 2

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8](6)=
//     @@ -2,8 +2,8 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "LastObjectSize": "341",
//     -        "ModTime": "0",
//     -        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//     +        "ModTime": "11",
//     +        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//              "RefCount": "1"
//          },
//          "Value": {
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:13](547)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key2"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "value2"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl/v0.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl/v0.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13",
//         "LastObjectSize": "547",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:12](344)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "LastObjectSize": "344",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl/v0.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "28ec6cc6d74185ca7bf2f597b583a79da3ba39e4",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13"
//         }
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:11](750)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key2"
//             }
//         },
//         {},
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl/v0.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "a423658b78ecdab66816cdc874814a85e7ef8612",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
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
//                     "ID": "gno.land/p/nt/avl/v0.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "d29a3290cc172ca5787fa65168e045bf8d5a3062",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "LastObjectSize": "750",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:10](343)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "LastObjectSize": "343",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl/v0.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "0b47217a25345d3884774f80c9d224b5aafe4aa8",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4](1)=
//     @@ -12,8 +12,8 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "55c53de4a2ad3643572f7acf760832c80fe8d65a",
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//     +                    "Hash": "8e34d7052c09cb70e9eec1ce9219bb6aa1104df5",
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10"
//                      },
//                      "Index": "0",
//                      "TV": null
//     @@ -23,7 +23,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "389",
//     -        "ModTime": "7",
//     +        "ModTime": "9",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
