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
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:7]={
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "RefCount": "0"
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
//                         "Hash": "9c2accff664935f02a9df69fa9ed2c2524fc3b5b",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
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
//                         "Hash": "aed377aa8f5da2daa19f3453d6104ef5ee34c5b8",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "RefCount": "0"
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
//                         "Hash": "bc070335da2de37db4dac38f976ea4ab19834bed",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:1"
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
//                         "Hash": "a42fd7516c704bd11c319d8bbcffdfa783864e25",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:0",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:0]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:0",
//         "ModTime": "4",
//         "RefCount": "0"
//     },
//     "Parent": null,
//     "SourceLoc": {
//         "File": "",
//         "Line": "0",
//         "PkgPath": ""
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
//                     "Hash": "3ed4d1261bfb4f0d40500bccc2b2e8af15197eb7",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                 },
//                 "FileName": "files/zrealm_avl1.go",
//                 "IsMethod": false,
//                 "Name": "init.0",
//                 "PkgPath": "gno.land/r/test",
//                 "SourceLoc": {
//                     "File": "files/zrealm_avl1.go",
//                     "Line": "10",
//                     "PkgPath": ""
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
//                     "Hash": "3ed4d1261bfb4f0d40500bccc2b2e8af15197eb7",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                 },
//                 "FileName": "files/zrealm_avl1.go",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "PkgPath": "gno.land/r/test",
//                 "SourceLoc": {
//                     "File": "files/zrealm_avl1.go",
//                     "Line": "15",
//                     "PkgPath": ""
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
//                         "Hash": "0e499b4767575949561a605ad0c85003f5068c5c",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                     }
//                 }
//             }
//         }
//     ]
// }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:2]
//
