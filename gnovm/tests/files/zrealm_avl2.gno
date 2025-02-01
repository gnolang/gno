// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/demo/avl"
)

var tree avl.Tree

func init() {
	tree.Set("key1", "value1")
}

func main() {
	var updated bool
	updated = tree.Set("key2", "value2")
	// println(tree, updated)
	println(updated, tree.Size())
}

// Output:
// false 2

// Realm:
// switchrealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "ModTime": "8",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
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
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "791bf9233a9e648e29ce5f1e9926e47e5c7ce4cb",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6"
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": true,
//                 "RefCount": "0"
//             },
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/avl.Node"
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
//                     "ID": "gno.land/p/demo/avl.Node"
//                 }
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
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
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
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "d50f58b5c1339b79c097d843c0b49cdd8e1f8e67",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10"
//         }
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:8]={
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
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "2"
//             },
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
//                     "Hash": "6a2c809421f492f491ea375a7d100665f1a4684c",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
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
//                     "Hash": "dd4a0ca3157011c0cf816d173a435eefb3d9ec88",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:7]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
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
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "060213e1d47505df4f6e5d4f987b5ac3fe69d52c",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3]={
//     "Fields": [
//         {
//             "AllocationInfo": {
//                 "AllocType": false,
//                 "AllocValue": false,
//                 "RefCount": "2"
//             },
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
//                     "Hash": "38517cfc3f06f9720f0bba6000ea1a97a7eae96c",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "ModTime": "6",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     }
// }
