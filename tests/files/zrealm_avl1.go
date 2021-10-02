// PKGPATH: gno.land/r/test
package test

import (
	"github.com/gnolang/gno/_test/avl"
)

var node *avl.Node

func init() {
	node = avl.NewNode("key0", "value0")
	node, _ = node.Set("key1", "value1")
}

func main() {
	var updated bool
	node, updated = node.Set("key2", "value2")
	//println(node, updated)
	println(updated, node.Size())
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
//                 "ID": "67c8d4035680e1de9f845da2566097cb688f34f1"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "67c8d4035680e1de9f845da2566097cb688f34f1"
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
//                 "ID": "67c8d4035680e1de9f845da2566097cb688f34f1"
//             },
//             "V": {
//                 "@type": "/gno.ptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.rft",
//                         "ID": "618d8d8164d24f0e1b132f437f88e2e20376d7ba"
//                     },
//                     "V": {
//                         "@type": "/gno.rfv",
//                         "Hash": "199ac12c488b7b89489bad87e34bf922dae55aaf",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "67c8d4035680e1de9f845da2566097cb688f34f1"
//             },
//             "V": {
//                 "@type": "/gno.ptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.rft",
//                         "ID": "618d8d8164d24f0e1b132f437f88e2e20376d7ba"
//                     },
//                     "V": {
//                         "@type": "/gno.rfv",
//                         "Hash": "022e1931e44efd15cfeb0fb7cc2f13aaf9430494",
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
//                 "ID": "67c8d4035680e1de9f845da2566097cb688f34f1"
//             },
//             "V": {
//                 "@type": "/gno.ptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.rft",
//                         "ID": "618d8d8164d24f0e1b132f437f88e2e20376d7ba"
//                     },
//                     "V": {
//                         "@type": "/gno.rfv",
//                         "Hash": "139c23df1b1da2f74db95f2165f33ac4aaab7004",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:1"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "67c8d4035680e1de9f845da2566097cb688f34f1"
//             },
//             "V": {
//                 "@type": "/gno.ptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.rft",
//                         "ID": "618d8d8164d24f0e1b132f437f88e2e20376d7ba"
//                     },
//                     "V": {
//                         "@type": "/gno.rfv",
//                         "Hash": "23e03266f0b1d56d7aa4097e3b91432d23ca6af6",
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
//                     "Hash": "86bfe249735ad7fb8388bdcb5779481eaa546c6f",
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
//                     "Hash": "86bfe249735ad7fb8388bdcb5779481eaa546c6f",
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
//                 "ID": "67c8d4035680e1de9f845da2566097cb688f34f1"
//             },
//             "V": {
//                 "@type": "/gno.ptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.rft",
//                         "ID": "618d8d8164d24f0e1b132f437f88e2e20376d7ba"
//                     },
//                     "V": {
//                         "@type": "/gno.rfv",
//                         "Hash": "2dee68a52f2534b4a014fb31f3fcea30e7b40fc4",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                     }
//                 }
//             }
//         }
//     ]
// }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:2]
//
