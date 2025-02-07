// PKGPATH: gno.land/r/test
package test

import (
	"github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl"
	"github.com/gnolang/gno/_test/timtadh/data_structures/types"
)

var tree *avl.AvlNode

func init() {
	tree, _ = tree.Put(types.String("key0"), "value0")
	tree, _ = tree.Put(types.String("key1"), "value1")
	tree, _ = tree.Put(types.String("key2"), "value2")
}

func main() {
	var updated bool
	tree, updated = tree.Put(types.String("key3"), "value3")
	println(updated, tree.Size())
}

// Output:
// false 4

// Realm:
// switchrealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:11]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/types.String"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key3"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "value3"
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             }
//         },
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:10]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "RefCount": "1"
//     },
//     "Value": {
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "2656aa76542f17ea5269bcba9c4239687d8a33c7",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:9]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/types.String"
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
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             }
//         },
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "1"
//             },
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "ec2a3363a32e5b693f5b0974f3641555fce052b8",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "ModTime": "9",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
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
//                 "value": "value0"
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "ModTime": "9",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "ModTime": "9",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     },
//     "Value": {
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "d7fbb234dca9f194f35fe5409a62db9daf39b0fc",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "ModTime": "9",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "RefCount": "1"
//     },
//     "Value": {
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "dfe919eaa94fc6b9b6c491e31f17356447a318b4",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/types.String"
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
//             "N": "AwAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "7"
//             },
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "194978fa0a568e3d821788a35a15c4035afd9be5",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "5"
//             },
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "a99545d51c903239d34ee4593851bfa38672f7ab",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "ModTime": "9",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:6]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "ModTime": "9",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "RefCount": "1"
//     },
//     "Value": {
//         "AllocationInfo": {
//             "AllocType": false,
//             "AllocValue": true,
//             "RefCount": "0"
//         },
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "4d388079817095a7ae3dca4ec1726a6bd595a054",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//         }
//     }
// }
