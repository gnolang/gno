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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0]=
//     @@ -1,8 +1,8 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0",
//     -        "ModTime": "0",
//     -        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//     +        "ModTime": "10",
//     +        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0",
//              "RefCount": "1"
//          },
//          "Value": {
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0]={
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
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0",
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
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0]={
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
//                     "Hash": "89bd18dbc22efecf5f7f989cb382d40f250b4bd1",
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
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
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
//             "Hash": "d0afd7713ecd9a6a5d78c6b4a220afddad7e84e7",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//     -        "ModTime": "6",
//     +        "ModTime": "8",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2:0",
//              "RefCount": "1"
//          },
//     @@ -17,8 +17,8 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "91da160d0fbd420f379c255e2f9d5ca91f4e72bc",
//     -                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0"
//     +                "Hash": "1e1c4c0ab622459cb4709ab0897b427df71ce931",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0"
//                  },
//                  "Index": "0",
//                  "TV": null
