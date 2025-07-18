// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/demo/avl"
)

var node *avl.Node

func init() {
	node = avl.NewNode("key0", "value0")
	// node, _ = node.Set("key0", "value0")
}

func main(cur realm) {
	var updated bool
	node, updated = node.Set("key1", "value1")
	// println(node, updated)
	println(updated, node.Size())
}

// Output:
// false 2

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0](6)=
//     @@ -2,8 +2,8 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0",
//              "LastObjectSize": "346",
//     -        "ModTime": "0",
//     -        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//     +        "ModTime": "10",
//     +        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0",
//              "RefCount": "1"
//          },
//          "Value": {
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0](549)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key1"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "value1"
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0",
//         "LastObjectSize": "549",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0](349)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0",
//         "LastObjectSize": "349",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "a977d58db6d61c23576437df4748214e49da042a",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0"
//         }
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0](757)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key1"
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
//                     "Hash": "50e6c85c948c0b0b3f8ceaa32f62c1f1dfb10c11",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0"
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
//                     "Hash": "407045b3a8228d72e258bd6aa1fcceabec22cb32",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0",
//         "LastObjectSize": "757",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0](347)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
//         "LastObjectSize": "347",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "36ed97472407a56dd19a141e34c18669c8d8e720",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//              "LastObjectSize": "397",
//     -        "ModTime": "6",
//     +        "ModTime": "8",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2:0",
//              "RefCount": "1"
//          },
//     @@ -18,8 +18,8 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "91da160d0fbd420f379c255e2f9d5ca91f4e72bc",
//     -                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0"
//     +                "Hash": "2d66d6728d3b7d2a59cbc5c59bad1be3db264449",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0"
//                  },
//                  "Index": "0",
//                  "TV": null
