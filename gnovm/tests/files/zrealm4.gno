// PKGPATH: gno.land/r/test
package test

import (
	"github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl"
	"github.com/gnolang/gno/_test/timtadh/data_structures/types"
)

var tree *avl.AvlNode

func init() {
	tree, _ = tree.Put(types.String("key0"), "value0")
}

func main() {
	var updated bool
	tree, updated = tree.Put(types.String("key0"), "value0-new")
	println(updated, tree.Size())
}

// Output:
// true 1

// Realm:
// switchrealm["gno.land/r/test"]
// u[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/types.String"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "value0-new"
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
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "ModTime": "5",
//         "OwnerID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "RefCount": "1"
//     }
// }
// u[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:4]={
//     "ObjectInfo": {
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "ModTime": "5",
//         "OwnerID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "6be1a84418e268a1867b4fadcbf952698c8cc08c",
//             "ObjectID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//         }
//     }
// }
