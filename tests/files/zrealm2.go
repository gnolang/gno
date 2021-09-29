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
//         {},
//         {}
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
//                     "Hash": "ccd4ec63e5c038ff545d5c4aaebd34dd23ae8e25",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2"
//                 },
//                 "FileName": "files/zrealm2.go",
//                 "IsMethod": false,
//                 "Name": "init.3",
//                 "PkgPath": "gno.land/r/test",
//                 "SourceLoc": {
//                     "File": "files/zrealm2.go",
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
//                 "ID": "0ba050da455a6aad7074eb2148d53ecd5becc26d"
//             },
//             "V": {
//                 "@type": "/gno.fun",
//                 "Closure": {
//                     "@type": "/gno.rfv",
//                     "Hash": "ccd4ec63e5c038ff545d5c4aaebd34dd23ae8e25",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2"
//                 },
//                 "FileName": "files/zrealm2.go",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "PkgPath": "gno.land/r/test",
//                 "SourceLoc": {
//                     "File": "files/zrealm2.go",
//                     "Line": "21",
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
//                 "ID": "ce75e799ed699fe6a487d6ca237759f5f203bee0"
//             },
//             "V": {
//                 "@type": "/gno.rfv",
//                 "Hash": "51ef9d8b69e49afb9509fcaa2e92e5db176cace7",
//                 "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//             }
//         }
//     ]
// }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:1]
