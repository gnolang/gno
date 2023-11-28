// PKGPATH: gno.land/r/test
package test

import (
	"github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl"
	"github.com/gnolang/gno/_test/timtadh/data-structures/types"
)

var tree *avl.AvlNode

func init() {
	tree, _ = tree.Put(types.String("key0"), "value0")
	tree, _ = tree.Put(types.String("key1"), "value1")
}

func main() {
	var updated bool
	tree, updated = tree.Put(types.String("key3"), "value3")
	println(updated, tree.Size())
}

// Output:
// false 3

// Realm:
// switchrealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:6]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/types.String"
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
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl.AvlNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl.AvlNode"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/types.String"
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
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl.AvlNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl.AvlNode"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl.AvlNode"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "61131413833a7638001260b674860d3d3afce0dc",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "ModTime": "5",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/types.String"
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
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl.AvlNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl.AvlNode"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl.AvlNode"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "bdc65e33d3ee6d6335e55f235115504922ee30fd",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "ModTime": "5",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "IsEscaped": true,
//         "ModTime": "5",
//         "RefCount": "2"
//     },
//     "Parent": null,
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "",
//             "Line": "0",
//             "Nonce": "0",
//             "PkgPath": "gno.land/r/test"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                 },
//                 "FileName": "main.gno",
//                 "IsMethod": false,
//                 "Name": "init.0",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "main.gno",
//                         "Line": "11",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                 },
//                 "FileName": "main.gno",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "main.gno",
//                         "Line": "16",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl.AvlNode"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl.AvlNode"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "5b4102ad6dc4a542b6b036ec6f5b99d0b37dbc64",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                     }
//                 }
//             }
//         }
//     ]
// }
