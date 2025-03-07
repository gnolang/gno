// PKGPATH: gno.land/r/test
package test

import (
	"std"
)

var node any

func init() {
	node = std.ChainHeight
}

func main() {
	// NOTE: this test uses ChainHeight and ChainID, which are "pure"
	// natively bound functions (ie. not indirections through a wrapper fn,
	// to convert the types to builtin go/gno identifiers).
	f := node.(func() int64)
	println(f())
	node = std.ChainID
	g := node.(func() string)
	println(g())
}

// Output:
// 123
// dev

// Realm:
// switchrealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "RefCount": "2"
//     },
//     "Parent": null,
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "0",
//             "File": "",
//             "Line": "0",
//             "PkgPath": "gno.land/r/test"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "IsClosure": false,
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PrimitiveType",
//                             "value": "16"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:10"
//                 },
//                 "FileName": "native.gno",
//                 "IsMethod": false,
//                 "Name": "ChainID",
//                 "NativeName": "ChainID",
//                 "NativePkg": "std",
//                 "PkgPath": "std",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "native.gno",
//                         "Line": "9",
//                         "PkgPath": "std"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "16"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "IsClosure": false,
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
//                 "FileName": "files/zrealm_natbind0.gno",
//                 "IsMethod": false,
//                 "Name": "init.1",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "files/zrealm_natbind0.gno",
//                         "Line": "10",
//                         "PkgPath": "gno.land/r/test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "IsClosure": false,
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
//                 "FileName": "files/zrealm_natbind0.gno",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/test",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "files/zrealm_natbind0.gno",
//                         "Line": "14",
//                         "PkgPath": "gno.land/r/test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [],
//                     "Results": []
//                 }
//             }
//         }
//     ]
// }
// u[a7f5397443359ea76c50be82c77f1f893a060925:10]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a7f5397443359ea76c50be82c77f1f893a060925:10",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "RefCount": "15"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:2"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "1",
//             "File": "native.gno",
//             "Line": "1",
//             "PkgPath": "std"
//         }
//     },
//     "Values": []
// }
