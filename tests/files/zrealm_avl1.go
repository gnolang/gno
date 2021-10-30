// PKGPATH: gno.land/r/test
package test

import (
	"github.com/gnolang/gno/_test/avl"
)

var tree *avl.Tree

func init() {
	tree = avl.NewTree("key0", "value0")
	tree, _ = tree.Set("key1", "value1")
}

func main() {
	var updated bool
	tree, updated = tree.Set("key2", "value2")
	//println(tree, updated)
	println(updated, tree.Size())
}

// Output:
// false 3

// Realm:
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:9]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "473287f8298dba7163a897908958f7c0eae733e2"
//             },
//             "V": {
//                 "@type": "/gno.st",
//                 "value": "key2"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "473287f8298dba7163a897908958f7c0eae733e2"
//             },
//             "V": {
//                 "@type": "/gno.st",
//                 "value": "value2"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "cb1525bced78da2c03c42fe15bf15663b584566e"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "6da88c34ba124c41f977db66a4fc5c1a951708d2"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "2734faa3f4c7500bdce18a66c4f013498a4c40d6"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "2734faa3f4c7500bdce18a66c4f013498a4c40d6"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:8]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "473287f8298dba7163a897908958f7c0eae733e2"
//             },
//             "V": {
//                 "@type": "/gno.st",
//                 "value": "key2"
//             }
//         },
//         {},
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "cb1525bced78da2c03c42fe15bf15663b584566e"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "6da88c34ba124c41f977db66a4fc5c1a951708d2"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "2734faa3f4c7500bdce18a66c4f013498a4c40d6"
//             },
//             "V": {
//                 "@type": "/gno.ptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.rft",
//                         "ID": "bacf836c433d04153744c70a6b8cf1f195b84588"
//                     },
//                     "V": {
//                         "@type": "/gno.rfv",
//                         "Hash": "3f9727638b2fd7934ab51db1823cd97c76fc66b2",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "2734faa3f4c7500bdce18a66c4f013498a4c40d6"
//             },
//             "V": {
//                 "@type": "/gno.ptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.rft",
//                         "ID": "bacf836c433d04153744c70a6b8cf1f195b84588"
//                     },
//                     "V": {
//                         "@type": "/gno.rfv",
//                         "Hash": "930ca20e7b61fa7878ce6acfa53fbe0a8e637eb4",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                     }
//                 }
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
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "473287f8298dba7163a897908958f7c0eae733e2"
//             },
//             "V": {
//                 "@type": "/gno.st",
//                 "value": "key1"
//             }
//         },
//         {},
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "cb1525bced78da2c03c42fe15bf15663b584566e"
//             }
//         },
//         {
//             "N": "AwAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "6da88c34ba124c41f977db66a4fc5c1a951708d2"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "2734faa3f4c7500bdce18a66c4f013498a4c40d6"
//             },
//             "V": {
//                 "@type": "/gno.ptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.rft",
//                         "ID": "bacf836c433d04153744c70a6b8cf1f195b84588"
//                     },
//                     "V": {
//                         "@type": "/gno.rfv",
//                         "Hash": "be18efd8b9972af09baf05008d2805d00b9aca85",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "2734faa3f4c7500bdce18a66c4f013498a4c40d6"
//             },
//             "V": {
//                 "@type": "/gno.ptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.rft",
//                         "ID": "bacf836c433d04153744c70a6b8cf1f195b84588"
//                     },
//                     "V": {
//                         "@type": "/gno.rfv",
//                         "Hash": "976b83782fbc30e6db881adf2e9ce652678a9d47",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "IsEscaped": true,
//         "ModTime": "6",
//         "RefCount": "2"
//     },
//     "Parent": null,
//     "Source": {
//         "@type": "/gno.rfn",
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
//                 "@type": "/gno.rft",
//                 "ID": "0ba050da455a6aad7074eb2148d53ecd5becc26d"
//             },
//             "V": {
//                 "@type": "/gno.fun",
//                 "Closure": {
//                     "@type": "/gno.rfv",
//                     "Escaped": true,
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                 },
//                 "FileName": "files/zrealm_avl1.go",
//                 "IsMethod": false,
//                 "Name": "init.0",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.rfn",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "files/zrealm_avl1.go",
//                         "Line": "10",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.rft",
//                     "ID": "0ba050da455a6aad7074eb2148d53ecd5becc26d"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "0ba050da455a6aad7074eb2148d53ecd5becc26d"
//             },
//             "V": {
//                 "@type": "/gno.fun",
//                 "Closure": {
//                     "@type": "/gno.rfv",
//                     "Escaped": true,
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                 },
//                 "FileName": "files/zrealm_avl1.go",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.rfn",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "files/zrealm_avl1.go",
//                         "Line": "15",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.rft",
//                     "ID": "0ba050da455a6aad7074eb2148d53ecd5becc26d"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "2734faa3f4c7500bdce18a66c4f013498a4c40d6"
//             },
//             "V": {
//                 "@type": "/gno.ptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.rft",
//                         "ID": "bacf836c433d04153744c70a6b8cf1f195b84588"
//                     },
//                     "V": {
//                         "@type": "/gno.rfv",
//                         "Hash": "0b09fb3942a9c60d9e1291da338660664383f15a",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                     }
//                 }
//             }
//         }
//     ]
// }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]
