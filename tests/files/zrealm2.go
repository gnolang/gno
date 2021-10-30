// PKGPATH: gno.land/r/test
package test

var root Node

type Node interface{}
type Key interface{}

type InnerNode struct {
	Key   Key
	Left  Node
	Right Node
}

func init() {
	root = InnerNode{
		Key: "old",
	}
}

func main() {
	root = InnerNode{
		Key: "new",
	}
}

// Realm:
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
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
//         {},
//         {}
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
//                     "ID": "ce75e799ed699fe6a487d6ca237759f5f203bee0"
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
//                 "FileName": "files/zrealm2.go",
//                 "IsMethod": false,
//                 "Name": "init.3",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.rfn",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "files/zrealm2.go",
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
//                 "ID": "0ba050da455a6aad7074eb2148d53ecd5becc26d"
//             },
//             "V": {
//                 "@type": "/gno.fun",
//                 "Closure": {
//                     "@type": "/gno.rfv",
//                     "Escaped": true,
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                 },
//                 "FileName": "files/zrealm2.go",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.rfn",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "files/zrealm2.go",
//                         "Line": "21",
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
//                 "ID": "ce75e799ed699fe6a487d6ca237759f5f203bee0"
//             },
//             "V": {
//                 "@type": "/gno.rfv",
//                 "Hash": "f99962d42e158b772fae4d18048b07b5907d9ca8",
//                 "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//             }
//         }
//     ]
// }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]
