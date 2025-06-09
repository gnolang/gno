// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/nt/avl"
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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7]=
//     @@ -1,8 +1,8 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//     -        "ModTime": "0",
//     -        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//     +        "ModTime": "10",
//     +        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "RefCount": "1"
//          },
//          "Value": {
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:12]={
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
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:11]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "d9eba6296483ae2f5f8cbfcdedc0b31b0d7458ae",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12"
//         }
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:10]={
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
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "c14fe6bc1ae08e8d8de9e8ca255e29eaa3faa8eb",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
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
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "7f06bab53f3d9a189dec22811bee43cc41fcdced",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:9]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "c8b27fe8b4c51dcefeac88c1046b28b6eff3a24a",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//     -        "ModTime": "6",
//     +        "ModTime": "8",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -17,8 +17,8 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "a6517cab33d5c4b6d2647a9df4ad7283b375a8f3",
//     -                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +                "Hash": "b25e30129a09cdc5d13e4ce76ae1b23ecab12200",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                  },
//                  "Index": "0",
//                  "TV": null
