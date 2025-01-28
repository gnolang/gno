// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/nt/avl"
)

var node *avl.Node

func init() {
	node = avl.NewNode("key0", "value0")
	node, _ = node.Set("key1", "value1")
}

func main() {
	var updated bool
	node, updated = node.Set("key2", "value2")
	// println(node, updated)
	println(updated, node.Size())
}

// Output:
// false 3

// Realm:
// switchrealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:6]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "ModTime": "11",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "cd518a6b18aac9ef53970759b00469f4feb4bd48",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "ModTime": "13",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "995cdf1dae0c0aa920f915fe6f1723027bace943",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//         }
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:15]={
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:15",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:14]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "4e4389eac32e83440a06885e2bdf7f9b21eb7f74",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:15"
//         }
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:13]={
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
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "89a6c669940efd4fd3753b210158430e80d17c68",
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
//                     "ID": "gno.land/p/nt/avl.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "65527e3570f1bcacd656ef37f8792681c76e2d26",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:12]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "0aca2c21e3c76e5522d526cb17d8947f72b257b0",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13"
//         }
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:11]={
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
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
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
//                     "Hash": "1a191527e1262b309b75e158375bf1a6e8470848",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6"
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
//                     "Hash": "b2e38511a31b0a58fa6d30d1d3034f0fa1438c75",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12"
//                 },
//                 "Index": "0",
//                 "TV": null
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
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "135056a38ec3764b01659804f933c2dc19a0de37",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11"
//         }
//     }
// }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]
