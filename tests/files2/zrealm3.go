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
// switchrealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.tpri",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.vstr",
//                 "value": "new"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/r/test.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/r/test.Node"
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
//         "@type": "/gno.nref",
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
//                 "@type": "/gno.ttyp"
//             },
//             "V": {
//                 "@type": "/gno.vtyp",
//                 "Type": {
//                     "@type": "/gno.tdec",
//                     "Base": {
//                         "@type": "/gno.tint",
//                         "Generic": "",
//                         "Methods": [],
//                         "PkgPath": "gno.land/r/test"
//                     },
//                     "Methods": [],
//                     "Name": "Key",
//                     "PkgPath": "gno.land/r/test"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.ttyp"
//             },
//             "V": {
//                 "@type": "/gno.vtyp",
//                 "Type": {
//                     "@type": "/gno.tdec",
//                     "Base": {
//                         "@type": "/gno.tstt",
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "Key",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.tref",
//                                     "ID": "gno.land/r/test.Key"
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "Left",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.tptr",
//                                     "Elt": {
//                                         "@type": "/gno.tref",
//                                         "ID": "gno.land/r/test.Node"
//                                     }
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "Right",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.tptr",
//                                     "Elt": {
//                                         "@type": "/gno.tref",
//                                         "ID": "gno.land/r/test.Node"
//                                     }
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/test"
//                     },
//                     "Methods": [],
//                     "Name": "Node",
//                     "PkgPath": "gno.land/r/test"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tfun",
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.vfun",
//                 "Closure": {
//                     "@type": "/gno.vref",
//                     "Escaped": true,
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                 },
//                 "FileName": "main.go",
//                 "IsMethod": false,
//                 "Name": "init.2",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.nref",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "main.go",
//                         "Line": "14",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.tfun",
//                     "Params": [],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tfun",
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.vfun",
//                 "Closure": {
//                     "@type": "/gno.vref",
//                     "Escaped": true,
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                 },
//                 "FileName": "main.go",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.nref",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "main.go",
//                         "Line": "20",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.tfun",
//                     "Params": [],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.tptr",
//                 "Elt": {
//                     "@type": "/gno.tref",
//                     "ID": "gno.land/r/test.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.vptr",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.tref",
//                         "ID": "gno.land/r/test.Node"
//                     },
//                     "V": {
//                         "@type": "/gno.vref",
//                         "Hash": "9d3f23dfb225ca4eb793e7c2ab3e15cb3621c180",
//                         "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                     }
//                 }
//             }
//         }
//     ]
// }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]
//
//
