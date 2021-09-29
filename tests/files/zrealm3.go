// PKGPATH: gno.land/r/test
package test

var root *Node

type Key interface{}

type Node struct {
	Key   Key
	Left  *Node
	Right *Node
}

func init() {
	root = &Node{
		Key: "old",
	}
}

func main() {
	root = &Node{
		Key: "new",
	}
}

// Realm:
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:3]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "473287f8298dba7163a897908958f7c0eae733e2"
//             },
//             "V": {
//                 "@type": "/gno.st",
//                 "value": "new"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "7b2e21e5a17ce618ada4860c549e3e24b9d73269"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "7b2e21e5a17ce618ada4860c549e3e24b9d73269"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:0",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:0]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:0",
//         "ModTime": "2",
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
//                 "ID": "1af40977153d0fabab9803bf33edeba8eb420cc5"
//             },
//             "V": {
//                 "@type": "/gno.typ",
//                 "Type": {
//                     "@type": "/gno.rft",
//                     "ID": "b06b716ff82d41a482d5c1cc3711002b74717639"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.rft",
//                 "ID": "1af40977153d0fabab9803bf33edeba8eb420cc5"
//             },
//             "V": {
//                 "@type": "/gno.typ",
//                 "Type": {
//                     "@type": "/gno.rft",
//                     "ID": "8f3fca65f6ca73d096c06f68e24ff93ea462d350"
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
//                     "Hash": "ccd4ec63e5c038ff545d5c4aaebd34dd23ae8e25",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2"
//                 },
//                 "FileName": "files/zrealm3.go",
//                 "IsMethod": false,
//                 "Name": "init.2",
//                 "PkgPath": "gno.land/r/test",
//                 "SourceLoc": {
//                     "File": "files/zrealm3.go",
//                     "Line": "14",
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
//                     "Hash": "ccd4ec63e5c038ff545d5c4aaebd34dd23ae8e25",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2"
//                 },
//                 "FileName": "files/zrealm3.go",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "PkgPath": "gno.land/r/test",
//                 "SourceLoc": {
//                     "File": "files/zrealm3.go",
//                     "Line": "20",
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
//                 "ID": "7b2e21e5a17ce618ada4860c549e3e24b9d73269"
//             },
//             "V": {
//                 "@type": "/gno.ptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.rft",
//                         "ID": "8f3fca65f6ca73d096c06f68e24ff93ea462d350"
//                     },
//                     "V": {
//                         "@type": "/gno.rfv",
//                         "Hash": "82801c38f53d7f4ca8f9d9c78bb319c434e2d6d8",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                     }
//                 }
//             }
//         }
//     ]
// }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:1]
