// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/demo/avl"
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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8](5)=
//     @@ -2,8 +2,8 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "LastObjectSize": "340",
//     -        "ModTime": "0",
//     -        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//     +        "ModTime": "9",
//     +        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "RefCount": "1"
//          },
//          "Value": {
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:11](545)={
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "LastObjectSize": "545",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:10](342)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "LastObjectSize": "342",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "c868579bb56defa78c588086d2569b998a0eae51",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11"
//         }
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:9](746)={
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "7d7b137abb9327e473f6130bf8468fd719183ac0",
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "22fab2a1d2444a6f768e41f161687d7337112d85",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "LastObjectSize": "746",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:8](340)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "LastObjectSize": "340",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "71b1ede447c3d696c3a17b52ea5f219b16a3365f",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -12,7 +12,7 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "38d92623e7ba4beb1d35693382b944d1dd9df1d2",
//     +                    "Hash": "98107500f95aab2e86e764076d6d44c289c8769f",
//                          "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//                      },
//                      "Index": "0",
