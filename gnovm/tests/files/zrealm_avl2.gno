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
// u[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
//     "ObjectInfo": {
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "ModTime": "8",
//         "OwnerID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "e221551f92f502d199b68084c3d5a30f072bb2e5",
//             "ObjectID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:6"
//         }
//     }
// }
// c[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:10]={
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
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "ModTime": "0",
//         "OwnerID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "RefCount": "1"
//     }
// }
// c[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:9]={
//     "ObjectInfo": {
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "ModTime": "0",
//         "OwnerID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "fca04d1bd566818259bd2a4fa74805c0215db218",
//             "ObjectID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:10"
//         }
//     }
// }
// c[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:8]={
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
//                     "Hash": "9dd79da898cdd530166a659a97d595392ee17c77",
//                     "ObjectID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
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
//                     "Hash": "1fba249e8d9732e0f85f59b9784a09b899137506",
//                     "ObjectID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "ModTime": "0",
//         "OwnerID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "RefCount": "1"
//     }
// }
// c[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:7]={
//     "ObjectInfo": {
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "ModTime": "0",
//         "OwnerID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "266b0530d93ebc972cb52c2e0d2ff486dc1d0f94",
//             "ObjectID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//         }
//     }
// }
// u[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:3]={
//     "Fields": [
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
//                     "Hash": "826fadf623de40be71afc9fa99fb05eb595adb2e",
//                     "ObjectID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "ModTime": "6",
//         "OwnerID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     }
// }
