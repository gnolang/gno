// PKGPATH: gno.land/r/test
package test

import (
	"github.com/gnolang/gno/_test/avl"
)

var tree *avl.Tree

func init() {
	tree = avl.NewTree("key0", "value0")
	// tree, _ = tree.Set("key0", "value0")
}

func main() {
	var updated bool
	tree, updated = tree.Set("key1", "value1")
	//println(tree, updated)
	println(updated, tree.Size())
}

// Output:
// false 2

// Realm:
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "473287f8298dba7163a897908958f7c0eae733e2"
//             },
//             "V": {
//                 "@type": "/gno.st",
//                 "value": "key0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "473287f8298dba7163a897908958f7c0eae733e2"
//             },
//             "V": {
//                 "@type": "/gno.st",
//                 "value": "value0"
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "ModTime": "5",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:6]={
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
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "473287f8298dba7163a897908958f7c0eae733e2"
//             },
//             "V": {
//                 "@type": "/gno.st",
//                 "value": "value1"
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
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
//                         "Hash": "5e8456c3f3ac84a752872124ae409356911f8012",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
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
//                         "Hash": "f76d625ff23ed8a5a0101f0fda2973bcc9f15e1e",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
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
//         "ModTime": "4",
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
//                 "FileName": "files/zrealm_avl0.go",
//                 "IsMethod": false,
//                 "Name": "init.0",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.rfn",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "files/zrealm_avl0.go",
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
//                 "FileName": "files/zrealm_avl0.go",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.rfn",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "files/zrealm_avl0.go",
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
//                         "Hash": "6eb8a7a0adbfabac4c3e4a4e257b0a043dc375bb",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                     }
//                 }
//             }
//         }
//     ]
// }
//
