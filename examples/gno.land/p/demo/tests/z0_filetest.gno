package main

import (
	ptests "gno.land/p/demo/tests"
	rtests "gno.land/r/demo/tests"
)

func main() {
	println(rtests.Counter())
	ptests.IncCounter()
	println(rtests.Counter())
}

// Output:
// 0
// 1

// Realm:
// switchrealm["errors"]
// c[be4bd56772770471b639574d919a92de16c0e522:1]={
//     "Block": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "be4bd56772770471b639574d919a92de16c0e522:2"
//     },
//     "FBlocks": [
//         {
//             "@type": "/gno.RefValue",
//             "Escaped": true,
//             "ObjectID": "be4bd56772770471b639574d919a92de16c0e522:3"
//         }
//     ],
//     "FNames": [
//         "errors.gno"
//     ],
//     "ObjectInfo": {
//         "ID": "be4bd56772770471b639574d919a92de16c0e522:1",
//         "ModTime": "0",
//         "RefCount": "1"
//     },
//     "PkgName": "errors",
//     "PkgPath": "errors"
// }
// c[be4bd56772770471b639574d919a92de16c0e522:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "be4bd56772770471b639574d919a92de16c0e522:2",
//         "IsEscaped": true,
//         "ModTime": "0",
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
//             "PkgPath": "errors"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.StructType",
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "s",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "errors"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "e",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "errors.errorString"
//                                             }
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "16"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "errors.gno",
//                                 "IsMethod": true,
//                                 "Name": "Error",
//                                 "PkgPath": "errors",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "errors.gno",
//                                         "Line": "67",
//                                         "Nonce": "0",
//                                         "PkgPath": "errors"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "e",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "errors.errorString"
//                                                 }
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "errorString",
//                     "PkgPath": "errors"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "text",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PrimitiveType",
//                             "value": "16"
//                         }
//                     }
//                 ],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": ".uverse.error"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "be4bd56772770471b639574d919a92de16c0e522:3"
//                 },
//                 "FileName": "errors.gno",
//                 "IsMethod": false,
//                 "Name": "New",
//                 "PkgPath": "errors",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "errors.gno",
//                         "Line": "58",
//                         "Nonce": "0",
//                         "PkgPath": "errors"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "text",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "16"
//                             }
//                         }
//                     ],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": ".uverse.error"
//                             }
//                         }
//                     ]
//                 }
//             }
//         }
//     ]
// }
// c[be4bd56772770471b639574d919a92de16c0e522:3]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "be4bd56772770471b639574d919a92de16c0e522:3",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "2"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "be4bd56772770471b639574d919a92de16c0e522:2"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "errors.gno",
//             "Line": "54",
//             "Nonce": "0",
//             "PkgPath": "errors"
//         }
//     },
//     "Values": []
// }
// switchrealm["strconv"]
// c[53ac87e8cb3cc9276228f74d38694a208cacb99b:3]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "53ac87e8cb3cc9276228f74d38694a208cacb99b:3",
//         "ModTime": "0",
//         "OwnerID": "53ac87e8cb3cc9276228f74d38694a208cacb99b:1",
//         "RefCount": "1"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "53ac87e8cb3cc9276228f74d38694a208cacb99b:2"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "strconv.gno",
//             "Line": "1",
//             "Nonce": "0",
//             "PkgPath": "strconv"
//         }
//     },
//     "Values": []
// }
// c[53ac87e8cb3cc9276228f74d38694a208cacb99b:1]={
//     "Block": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "53ac87e8cb3cc9276228f74d38694a208cacb99b:2"
//     },
//     "FBlocks": [
//         {
//             "@type": "/gno.RefValue",
//             "Hash": "6fcf65e36fc04c3393589f67313652f8e4bb7593",
//             "ObjectID": "53ac87e8cb3cc9276228f74d38694a208cacb99b:3"
//         }
//     ],
//     "FNames": [
//         "strconv.gno"
//     ],
//     "ObjectInfo": {
//         "ID": "53ac87e8cb3cc9276228f74d38694a208cacb99b:1",
//         "ModTime": "0",
//         "RefCount": "1"
//     },
//     "PkgName": "strconv",
//     "PkgPath": "strconv"
// }
// c[53ac87e8cb3cc9276228f74d38694a208cacb99b:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "53ac87e8cb3cc9276228f74d38694a208cacb99b:2",
//         "IsEscaped": true,
//         "ModTime": "0",
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
//             "PkgPath": "strconv"
//         }
//     },
//     "Values": []
// }
// switchrealm["std"]
// c[a7f5397443359ea76c50be82c77f1f893a060925:4]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a7f5397443359ea76c50be82c77f1f893a060925:4",
//         "ModTime": "0",
//         "OwnerID": "a7f5397443359ea76c50be82c77f1f893a060925:1",
//         "RefCount": "1"
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
//             "File": "banker.gno",
//             "Line": "1",
//             "Nonce": "0",
//             "PkgPath": "std"
//         }
//     },
//     "Values": []
// }
// c[a7f5397443359ea76c50be82c77f1f893a060925:5]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a7f5397443359ea76c50be82c77f1f893a060925:5",
//         "ModTime": "0",
//         "OwnerID": "a7f5397443359ea76c50be82c77f1f893a060925:1",
//         "RefCount": "1"
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
//             "File": "coins.gno",
//             "Line": "1",
//             "Nonce": "0",
//             "PkgPath": "std"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.PackageType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "PkgPath": "strconv"
//             }
//         }
//     ]
// }
// c[a7f5397443359ea76c50be82c77f1f893a060925:6]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a7f5397443359ea76c50be82c77f1f893a060925:6",
//         "ModTime": "0",
//         "OwnerID": "a7f5397443359ea76c50be82c77f1f893a060925:1",
//         "RefCount": "1"
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
//             "File": "context.gno",
//             "Line": "1",
//             "Nonce": "0",
//             "PkgPath": "std"
//         }
//     },
//     "Values": []
// }
// c[a7f5397443359ea76c50be82c77f1f893a060925:7]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a7f5397443359ea76c50be82c77f1f893a060925:7",
//         "ModTime": "0",
//         "OwnerID": "a7f5397443359ea76c50be82c77f1f893a060925:1",
//         "RefCount": "1"
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
//             "File": "crypto.gno",
//             "Line": "1",
//             "Nonce": "0",
//             "PkgPath": "std"
//         }
//     },
//     "Values": []
// }
// c[a7f5397443359ea76c50be82c77f1f893a060925:8]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a7f5397443359ea76c50be82c77f1f893a060925:8",
//         "ModTime": "0",
//         "OwnerID": "a7f5397443359ea76c50be82c77f1f893a060925:1",
//         "RefCount": "1"
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
//             "File": "frame.gno",
//             "Line": "1",
//             "Nonce": "0",
//             "PkgPath": "std"
//         }
//     },
//     "Values": []
// }
// c[a7f5397443359ea76c50be82c77f1f893a060925:1]={
//     "Block": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:2"
//     },
//     "FBlocks": [
//         {
//             "@type": "/gno.RefValue",
//             "Escaped": true,
//             "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:3"
//         },
//         {
//             "@type": "/gno.RefValue",
//             "Hash": "a7830cc99140c70927a298b51b0b9d1c8b958d49",
//             "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:4"
//         },
//         {
//             "@type": "/gno.RefValue",
//             "Hash": "67c04c9244a6a9566a77ddd688a6d9d8541e1066",
//             "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:5"
//         },
//         {
//             "@type": "/gno.RefValue",
//             "Hash": "04511f43e4a0a86b6155efa6d9283cd8f26f9507",
//             "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:6"
//         },
//         {
//             "@type": "/gno.RefValue",
//             "Hash": "014556803a98a074a5ff65a3c69a6637a56e2cb8",
//             "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:7"
//         },
//         {
//             "@type": "/gno.RefValue",
//             "Hash": "0731dc7a780f89dea11e5e367c0ee1287155256a",
//             "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:8"
//         }
//     ],
//     "FNames": [
//         "addr_set.gno",
//         "banker.gno",
//         "coins.gno",
//         "context.gno",
//         "crypto.gno",
//         "frame.gno"
//     ],
//     "ObjectInfo": {
//         "ID": "a7f5397443359ea76c50be82c77f1f893a060925:1",
//         "ModTime": "0",
//         "RefCount": "1"
//     },
//     "PkgName": "std",
//     "PkgPath": "std"
// }
// c[a7f5397443359ea76c50be82c77f1f893a060925:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a7f5397443359ea76c50be82c77f1f893a060925:2",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "7"
//     },
//     "Parent": null,
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "",
//             "Line": "0",
//             "Nonce": "0",
//             "PkgPath": "std"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.InterfaceType",
//                         "Generic": "",
//                         "Methods": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "Size",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "32"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "AddAddress",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": ".arg_0",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": ".uverse.error"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "HasAddress",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": ".arg_0",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "4"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         ],
//                         "PkgPath": "std"
//                     },
//                     "Methods": [],
//                     "Name": "AddressSet",
//                     "PkgPath": "std"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.PrimitiveType",
//                         "value": "16"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "a",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Address"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "16"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "crypto.gno",
//                                 "IsMethod": true,
//                                 "Name": "String",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "crypto.gno",
//                                         "Line": "5",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "a",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "Address",
//                     "PkgPath": "std"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.SliceType",
//                         "Elt": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Address"
//                         },
//                         "Vrd": false
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "alist",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.AddressList"
//                                             }
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "32"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "addr_set.gno",
//                                 "IsMethod": true,
//                                 "Name": "Size",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "addr_set.gno",
//                                         "Line": "24",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "alist",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "std.AddressList"
//                                                 }
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "32"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         },
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "alist",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.AddressList"
//                                             }
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "newAddr",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Address"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": ".uverse.error"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "addr_set.gno",
//                                 "IsMethod": true,
//                                 "Name": "AddAddress",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "addr_set.gno",
//                                         "Line": "28",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "alist",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "std.AddressList"
//                                                 }
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "newAddr",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": ".uverse.error"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         },
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "alist",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.AddressList"
//                                             }
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "newAddr",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Address"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "4"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "addr_set.gno",
//                                 "IsMethod": true,
//                                 "Name": "HasAddress",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "addr_set.gno",
//                                         "Line": "39",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "alist",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "std.AddressList"
//                                                 }
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "newAddr",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "4"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "AddressList",
//                     "PkgPath": "std"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.InterfaceType",
//                         "Generic": "",
//                         "Methods": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "GetCoins",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "addr",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "dst",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Coins"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "SendCoins",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "from",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "to",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "amt",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Coins"
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "TotalCoin",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "denom",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "1024"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "IssueCoin",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "addr",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "denom",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "amount",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "1024"
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "RemoveCoin",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "addr",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "denom",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "amount",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "1024"
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             }
//                         ],
//                         "PkgPath": "std"
//                     },
//                     "Methods": [],
//                     "Name": "Banker",
//                     "PkgPath": "std"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.SliceType",
//                         "Elt": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Coin"
//                         },
//                         "Vrd": false
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "cz",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Coins"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "16"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "coins.gno",
//                                 "IsMethod": true,
//                                 "Name": "String",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "coins.gno",
//                                         "Line": "29",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "cz",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Coins"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         },
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "cz",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Coins"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "denom",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "16"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "1024"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "coins.gno",
//                                 "IsMethod": true,
//                                 "Name": "AmountOf",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "coins.gno",
//                                         "Line": "40",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "cz",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Coins"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "denom",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "1024"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         },
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "a",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Coins"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "b",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Coins"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Coins"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "coins.gno",
//                                 "IsMethod": true,
//                                 "Name": "Add",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "coins.gno",
//                                         "Line": "49",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "a",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Coins"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "b",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Coins"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Coins"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "Coins",
//                     "PkgPath": "std"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.StructType",
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "Denom",
//                                 "Tag": "json:\"denom\"",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "Amount",
//                                 "Tag": "json:\"amount\"",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "1024"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "std"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "c",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Coin"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "16"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "coins.gno",
//                                 "IsMethod": true,
//                                 "Name": "String",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "coins.gno",
//                                         "Line": "15",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "c",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Coin"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         },
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "c",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Coin"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "other",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Coin"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "4"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "coins.gno",
//                                 "IsMethod": true,
//                                 "Name": "IsGTE",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "coins.gno",
//                                         "Line": "19",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "c",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Coin"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "other",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Coin"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "4"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "Coin",
//                     "PkgPath": "std"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.PrimitiveType",
//                         "value": "4096"
//                     },
//                     "Methods": [],
//                     "Name": "BankerType",
//                     "PkgPath": "std"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.StructType",
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "nativeBanker",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "std.Banker"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "std"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "ba",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.bankAdapter"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "addr",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Address"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "dst",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Coins"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "banker.gno",
//                                 "IsMethod": true,
//                                 "Name": "GetCoins",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "banker.gno",
//                                         "Line": "47",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "ba",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.bankAdapter"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "addr",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "dst",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Coins"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         },
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "ba",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.bankAdapter"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "from",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Address"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "to",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Address"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "amt",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Coins"
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "banker.gno",
//                                 "IsMethod": true,
//                                 "Name": "SendCoins",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "banker.gno",
//                                         "Line": "56",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "ba",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.bankAdapter"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "from",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "to",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "amt",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Coins"
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             }
//                         },
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "ba",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.bankAdapter"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "denom",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "16"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "1024"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "banker.gno",
//                                 "IsMethod": true,
//                                 "Name": "TotalCoin",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "banker.gno",
//                                         "Line": "60",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "ba",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.bankAdapter"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "denom",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "1024"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         },
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "ba",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.bankAdapter"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "addr",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Address"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "denom",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "16"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "amount",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "1024"
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "banker.gno",
//                                 "IsMethod": true,
//                                 "Name": "IssueCoin",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "banker.gno",
//                                         "Line": "64",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "ba",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.bankAdapter"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "addr",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "denom",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "amount",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "1024"
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             }
//                         },
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "ba",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.bankAdapter"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "addr",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Address"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "denom",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "16"
//                                         }
//                                     },
//                                     {
//                                         "Embedded": false,
//                                         "Name": "amount",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "1024"
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "banker.gno",
//                                 "IsMethod": true,
//                                 "Name": "RemoveCoin",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "banker.gno",
//                                         "Line": "68",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "ba",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.bankAdapter"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "addr",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "denom",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         },
//                                         {
//                                             "Embedded": false,
//                                             "Name": "amount",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "1024"
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "bankAdapter",
//                     "PkgPath": "std"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.ArrayType",
//                         "Elt": {
//                             "@type": "/gno.PrimitiveType",
//                             "value": "4096"
//                         },
//                         "Len": "20",
//                         "Vrd": false
//                     },
//                     "Methods": [],
//                     "Name": "RawAddress",
//                     "PkgPath": "std"
//                 }
//             }
//         },
//         {
//             "N": "FAAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.StructType",
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "addr",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "std.Address"
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "pkgPath",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "std"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "r",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Realm"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Address"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "frame.gno",
//                                 "IsMethod": true,
//                                 "Name": "Addr",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "frame.gno",
//                                         "Line": "8",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "r",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Realm"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Address"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         },
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "r",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Realm"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "16"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "frame.gno",
//                                 "IsMethod": true,
//                                 "Name": "PkgPath",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "frame.gno",
//                                         "Line": "12",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "r",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Realm"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         },
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "r",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "std.Realm"
//                                         }
//                                     }
//                                 ],
//                                 "Results": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "4"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "frame.gno",
//                                 "IsMethod": true,
//                                 "Name": "IsUser",
//                                 "PkgPath": "std",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "frame.gno",
//                                         "Line": "16",
//                                         "Nonce": "0",
//                                         "PkgPath": "std"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "r",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "std.Realm"
//                                             }
//                                         }
//                                     ],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "4"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "Realm",
//                     "PkgPath": "std"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "std.AddressList"
//                             }
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "a7f5397443359ea76c50be82c77f1f893a060925:3"
//                 },
//                 "FileName": "addr_set.gno",
//                 "IsMethod": false,
//                 "Name": "NewAddressList",
//                 "PkgPath": "std",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "addr_set.gno",
//                         "Line": "20",
//                         "Nonce": "0",
//                         "PkgPath": "std"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "std.AddressList"
//                                 }
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "std.BankerType"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "std.BankerType"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "std.BankerType"
//             }
//         },
//         {
//             "N": "AwAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "std.BankerType"
//             }
//         }
//     ]
// }
// c[a7f5397443359ea76c50be82c77f1f893a060925:3]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "a7f5397443359ea76c50be82c77f1f893a060925:3",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "2"
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
//             "File": "addr_set.gno",
//             "Line": "1",
//             "Nonce": "0",
//             "PkgPath": "std"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.PackageType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "PkgPath": "errors"
//             }
//         }
//     ]
// }
// switchrealm["gno.land/p/demo/tests/subtests"]
// c[0d4452d3618819945f654f2f269956e7743ee18b:1]={
//     "Block": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "0d4452d3618819945f654f2f269956e7743ee18b:2"
//     },
//     "FBlocks": [
//         {
//             "@type": "/gno.RefValue",
//             "Escaped": true,
//             "ObjectID": "0d4452d3618819945f654f2f269956e7743ee18b:3"
//         }
//     ],
//     "FNames": [
//         "subtests.gno"
//     ],
//     "ObjectInfo": {
//         "ID": "0d4452d3618819945f654f2f269956e7743ee18b:1",
//         "ModTime": "0",
//         "RefCount": "1"
//     },
//     "PkgName": "subtests",
//     "PkgPath": "gno.land/p/demo/tests/subtests"
// }
// c[0d4452d3618819945f654f2f269956e7743ee18b:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "0d4452d3618819945f654f2f269956e7743ee18b:2",
//         "IsEscaped": true,
//         "ModTime": "0",
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
//             "PkgPath": "gno.land/p/demo/tests/subtests"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Realm"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0d4452d3618819945f654f2f269956e7743ee18b:3"
//                 },
//                 "FileName": "subtests.gno",
//                 "IsMethod": false,
//                 "Name": "GetCurrentRealm",
//                 "PkgPath": "gno.land/p/demo/tests/subtests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "subtests.gno",
//                         "Line": "7",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests/subtests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "std.Realm"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Realm"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0d4452d3618819945f654f2f269956e7743ee18b:3"
//                 },
//                 "FileName": "subtests.gno",
//                 "IsMethod": false,
//                 "Name": "GetPrevRealm",
//                 "PkgPath": "gno.land/p/demo/tests/subtests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "subtests.gno",
//                         "Line": "11",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests/subtests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "std.Realm"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "fn",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [],
//                             "Results": []
//                         }
//                     }
//                 ],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0d4452d3618819945f654f2f269956e7743ee18b:3"
//                 },
//                 "FileName": "subtests.gno",
//                 "IsMethod": false,
//                 "Name": "Exec",
//                 "PkgPath": "gno.land/p/demo/tests/subtests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "subtests.gno",
//                         "Line": "15",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests/subtests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "fn",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [],
//                                 "Results": []
//                             }
//                         }
//                     ],
//                     "Results": []
//                 }
//             }
//         }
//     ]
// }
// c[0d4452d3618819945f654f2f269956e7743ee18b:3]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "0d4452d3618819945f654f2f269956e7743ee18b:3",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "4"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "0d4452d3618819945f654f2f269956e7743ee18b:2"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "subtests.gno",
//             "Line": "1",
//             "Nonce": "0",
//             "PkgPath": "gno.land/p/demo/tests/subtests"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.PackageType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "PkgPath": "std"
//             }
//         }
//     ]
// }
// switchrealm["gno.land/r/demo/tests/subtests"]
// c[fc6bed79a4ac3aef903c43415d367c10e1ea9871:1]={
//     "Block": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "fc6bed79a4ac3aef903c43415d367c10e1ea9871:2"
//     },
//     "FBlocks": [
//         {
//             "@type": "/gno.RefValue",
//             "Escaped": true,
//             "ObjectID": "fc6bed79a4ac3aef903c43415d367c10e1ea9871:3"
//         }
//     ],
//     "FNames": [
//         "subtests.gno"
//     ],
//     "ObjectInfo": {
//         "ID": "fc6bed79a4ac3aef903c43415d367c10e1ea9871:1",
//         "ModTime": "0",
//         "RefCount": "1"
//     },
//     "PkgName": "subtests",
//     "PkgPath": "gno.land/r/demo/tests/subtests"
// }
// c[fc6bed79a4ac3aef903c43415d367c10e1ea9871:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "fc6bed79a4ac3aef903c43415d367c10e1ea9871:2",
//         "IsEscaped": true,
//         "ModTime": "0",
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
//             "PkgPath": "gno.land/r/demo/tests/subtests"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Realm"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "fc6bed79a4ac3aef903c43415d367c10e1ea9871:3"
//                 },
//                 "FileName": "subtests.gno",
//                 "IsMethod": false,
//                 "Name": "GetCurrentRealm",
//                 "PkgPath": "gno.land/r/demo/tests/subtests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "subtests.gno",
//                         "Line": "7",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests/subtests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "std.Realm"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Realm"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "fc6bed79a4ac3aef903c43415d367c10e1ea9871:3"
//                 },
//                 "FileName": "subtests.gno",
//                 "IsMethod": false,
//                 "Name": "GetPrevRealm",
//                 "PkgPath": "gno.land/r/demo/tests/subtests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "subtests.gno",
//                         "Line": "11",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests/subtests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "std.Realm"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "fn",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [],
//                             "Results": []
//                         }
//                     }
//                 ],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "fc6bed79a4ac3aef903c43415d367c10e1ea9871:3"
//                 },
//                 "FileName": "subtests.gno",
//                 "IsMethod": false,
//                 "Name": "Exec",
//                 "PkgPath": "gno.land/r/demo/tests/subtests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "subtests.gno",
//                         "Line": "15",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests/subtests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "fn",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [],
//                                 "Results": []
//                             }
//                         }
//                     ],
//                     "Results": []
//                 }
//             }
//         }
//     ]
// }
// c[fc6bed79a4ac3aef903c43415d367c10e1ea9871:3]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "fc6bed79a4ac3aef903c43415d367c10e1ea9871:3",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "4"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "fc6bed79a4ac3aef903c43415d367c10e1ea9871:2"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "subtests.gno",
//             "Line": "1",
//             "Nonce": "0",
//             "PkgPath": "gno.land/r/demo/tests/subtests"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.PackageType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "PkgPath": "std"
//             }
//         }
//     ]
// }
// switchrealm["gno.land/r/demo/tests"]
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:1]={
//     "Block": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2"
//     },
//     "FBlocks": [
//         {
//             "@type": "/gno.RefValue",
//             "Escaped": true,
//             "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3"
//         },
//         {
//             "@type": "/gno.RefValue",
//             "Escaped": true,
//             "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//         }
//     ],
//     "FNames": [
//         "interfaces.gno",
//         "tests.gno"
//     ],
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:1",
//         "ModTime": "0",
//         "RefCount": "1"
//     },
//     "PkgName": "tests",
//     "PkgPath": "gno.land/r/demo/tests"
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "3"
//     },
//     "Parent": null,
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "",
//             "Line": "0",
//             "Nonce": "0",
//             "PkgPath": "gno.land/r/demo/tests"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.InterfaceType",
//                         "Generic": "",
//                         "Methods": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "String",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests"
//                     },
//                     "Methods": [],
//                     "Name": "Stringer",
//                     "PkgPath": "gno.land/r/demo/tests"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.StructType",
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "Field",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "t",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests.TestRealmObject"
//                                             }
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "tests.gno",
//                                 "IsMethod": true,
//                                 "Name": "Modify",
//                                 "PkgPath": "gno.land/r/demo/tests",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "tests.gno",
//                                         "Line": "42",
//                                         "Nonce": "0",
//                                         "PkgPath": "gno.land/r/demo/tests"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "t",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/demo/tests.TestRealmObject"
//                                                 }
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "TestRealmObject",
//                     "PkgPath": "gno.land/r/demo/tests"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.StructType",
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "Name",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "Child",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PointerType",
//                                     "Elt": {
//                                         "@type": "/gno.RefType",
//                                         "ID": "gno.land/r/demo/tests.TestNode"
//                                     }
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests"
//                     },
//                     "Methods": [],
//                     "Name": "TestNode",
//                     "PkgPath": "gno.land/r/demo/tests"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "str",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests.Stringer"
//                         }
//                     }
//                 ],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3"
//                 },
//                 "FileName": "interfaces.gno",
//                 "IsMethod": false,
//                 "Name": "AddStringer",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "interfaces.gno",
//                         "Line": "13",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "str",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests.Stringer"
//                             }
//                         }
//                     ],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "path",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PrimitiveType",
//                             "value": "16"
//                         }
//                     }
//                 ],
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3"
//                 },
//                 "FileName": "interfaces.gno",
//                 "IsMethod": false,
//                 "Name": "Render",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "interfaces.gno",
//                         "Line": "20",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "path",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "16"
//                             }
//                         }
//                     ],
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
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "IncCounter",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "11",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PrimitiveType",
//                             "value": "32"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "Counter",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "15",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "32"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "CurrentRealmPath",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "19",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
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
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "AssertOriginCall",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "23",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PrimitiveType",
//                             "value": "4"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "IsOriginCall",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "27",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "4"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "t",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests.TestRealmObject"
//                             }
//                         }
//                     }
//                 ],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "ModifyTestRealmObject",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "38",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "t",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests.TestRealmObject"
//                                 }
//                             }
//                         }
//                     ],
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "InitTestNodes",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "60",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "ModTestNodes",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "65",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "PrintTestNodes",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "73",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Realm"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "GetPrevRealm",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "77",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "std.Realm"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Realm"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "GetRSubtestsPrevRealm",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "81",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "std.Realm"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "fn",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [],
//                             "Results": []
//                         }
//                     }
//                 ],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "Exec",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "85",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "fn",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [],
//                                 "Results": []
//                             }
//                         }
//                     ],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests.Stringer"
//                 },
//                 "Vrd": false
//             }
//         },
//         {
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
//                     "ID": "gno.land/r/demo/tests.TestNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests.TestNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests.TestNode"
//                 }
//             }
//         }
//     ]
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "3"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "interfaces.gno",
//             "Line": "1",
//             "Nonce": "0",
//             "PkgPath": "gno.land/r/demo/tests"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.PackageType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "PkgPath": "strconv"
//             }
//         }
//     ]
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "13"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "tests.gno",
//             "Line": "1",
//             "Nonce": "0",
//             "PkgPath": "gno.land/r/demo/tests"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.PackageType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "PkgPath": "std"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PackageType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "PkgPath": "gno.land/r/demo/tests/subtests"
//             }
//         }
//     ]
// }
// switchrealm["gno.land/p/demo/tests"]
// c[b609759cdf9e3b49753fd5632a02365700aff2b6:1]={
//     "Block": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:2"
//     },
//     "FBlocks": [
//         {
//             "@type": "/gno.RefValue",
//             "Escaped": true,
//             "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:3"
//         }
//     ],
//     "FNames": [
//         "tests.gno"
//     ],
//     "ObjectInfo": {
//         "ID": "b609759cdf9e3b49753fd5632a02365700aff2b6:1",
//         "ModTime": "0",
//         "RefCount": "1"
//     },
//     "PkgName": "tests",
//     "PkgPath": "gno.land/p/demo/tests"
// }
// c[b609759cdf9e3b49753fd5632a02365700aff2b6:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "init"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "b609759cdf9e3b49753fd5632a02365700aff2b6:4",
//         "ModTime": "0",
//         "OwnerID": "b609759cdf9e3b49753fd5632a02365700aff2b6:2",
//         "RefCount": "1"
//     }
// }
// c[b609759cdf9e3b49753fd5632a02365700aff2b6:5]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "init"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "b609759cdf9e3b49753fd5632a02365700aff2b6:5",
//         "ModTime": "0",
//         "OwnerID": "b609759cdf9e3b49753fd5632a02365700aff2b6:2",
//         "RefCount": "1"
//     }
// }
// c[b609759cdf9e3b49753fd5632a02365700aff2b6:6]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "init"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "b609759cdf9e3b49753fd5632a02365700aff2b6:6",
//         "ModTime": "0",
//         "OwnerID": "b609759cdf9e3b49753fd5632a02365700aff2b6:2",
//         "RefCount": "1"
//     }
// }
// c[b609759cdf9e3b49753fd5632a02365700aff2b6:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "b609759cdf9e3b49753fd5632a02365700aff2b6:2",
//         "IsEscaped": true,
//         "ModTime": "0",
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
//             "PkgPath": "gno.land/p/demo/tests"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.StructType",
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "Field",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/p/demo/tests"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "o2",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/p/demo/tests.TestRealmObject2"
//                                             }
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "tests.gno",
//                                 "IsMethod": true,
//                                 "Name": "Modify",
//                                 "PkgPath": "gno.land/p/demo/tests",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "tests.gno",
//                                         "Line": "29",
//                                         "Nonce": "0",
//                                         "PkgPath": "gno.land/p/demo/tests"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "o2",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/p/demo/tests.TestRealmObject2"
//                                                 }
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "TestRealmObject2",
//                     "PkgPath": "gno.land/p/demo/tests"
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
//                     "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:3"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "IncCounter",
//                 "PkgPath": "gno.land/p/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "14",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests"
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
//                     "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:3"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "CurrentRealmPath",
//                 "PkgPath": "gno.land/p/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "18",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
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
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:3"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "init.3",
//                 "PkgPath": "gno.land/p/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "39",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests"
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
//                     "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:3"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "ModifyTestRealmObject2a",
//                 "PkgPath": "gno.land/p/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "45",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests"
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
//                     "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:3"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "ModifyTestRealmObject2b",
//                 "PkgPath": "gno.land/p/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "49",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests"
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
//                     "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:3"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "ModifyTestRealmObject2c",
//                 "PkgPath": "gno.land/p/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "53",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests"
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Realm"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:3"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "GetPrevRealm",
//                 "PkgPath": "gno.land/p/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "57",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "std.Realm"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Realm"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:3"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "GetPSubtestsPrevRealm",
//                 "PkgPath": "gno.land/p/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "61",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "std.Realm"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Realm"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:3"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "GetRTestsGetPrevRealm",
//                 "PkgPath": "gno.land/p/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "65",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "std.Realm"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "fn",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [],
//                             "Results": []
//                         }
//                     }
//                 ],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:3"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "Exec",
//                 "PkgPath": "gno.land/p/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "70",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/p/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "fn",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [],
//                                 "Results": []
//                             }
//                         }
//                     ],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/tests.TestRealmObject2"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "79dc2749013d6a945f7ebe251518b3b69a9cc52c",
//                 "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:4"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/p/demo/tests.TestRealmObject2"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "1c2bee184883fcd29e4612de54206e4f0f60edd9",
//                 "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:5"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/tests.TestRealmObject2"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/p/demo/tests.TestRealmObject2"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "22d53dd8598ce1478e255929308ea4750d2004bc",
//                         "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:6"
//                     }
//                 }
//             }
//         }
//     ]
// }
// c[b609759cdf9e3b49753fd5632a02365700aff2b6:3]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "b609759cdf9e3b49753fd5632a02365700aff2b6:3",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "11"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "b609759cdf9e3b49753fd5632a02365700aff2b6:2"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "tests.gno",
//             "Line": "1",
//             "Nonce": "0",
//             "PkgPath": "gno.land/p/demo/tests"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.PackageType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "PkgPath": "std"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PackageType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "PkgPath": "gno.land/p/demo/tests/subtests"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PackageType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "PkgPath": "gno.land/r/demo/tests"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PackageType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "PkgPath": "gno.land/r/demo/tests"
//             }
//         }
//     ]
// }
// switchrealm["gno.land/r/demo/tests"]
// switchrealm["gno.land/r/demo/tests"]
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2",
//         "IsEscaped": true,
//         "ModTime": "4",
//         "RefCount": "3"
//     },
//     "Parent": null,
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "",
//             "Line": "0",
//             "Nonce": "0",
//             "PkgPath": "gno.land/r/demo/tests"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.InterfaceType",
//                         "Generic": "",
//                         "Methods": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "String",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [],
//                                     "Results": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "16"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests"
//                     },
//                     "Methods": [],
//                     "Name": "Stringer",
//                     "PkgPath": "gno.land/r/demo/tests"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.StructType",
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "Field",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "t",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests.TestRealmObject"
//                                             }
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "tests.gno",
//                                 "IsMethod": true,
//                                 "Name": "Modify",
//                                 "PkgPath": "gno.land/r/demo/tests",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "tests.gno",
//                                         "Line": "42",
//                                         "Nonce": "0",
//                                         "PkgPath": "gno.land/r/demo/tests"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "t",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/demo/tests.TestRealmObject"
//                                                 }
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "TestRealmObject",
//                     "PkgPath": "gno.land/r/demo/tests"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.StructType",
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "Name",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "Child",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PointerType",
//                                     "Elt": {
//                                         "@type": "/gno.RefType",
//                                         "ID": "gno.land/r/demo/tests.TestNode"
//                                     }
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests"
//                     },
//                     "Methods": [],
//                     "Name": "TestNode",
//                     "PkgPath": "gno.land/r/demo/tests"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "str",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests.Stringer"
//                         }
//                     }
//                 ],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3"
//                 },
//                 "FileName": "interfaces.gno",
//                 "IsMethod": false,
//                 "Name": "AddStringer",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "interfaces.gno",
//                         "Line": "13",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "str",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests.Stringer"
//                             }
//                         }
//                     ],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "path",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PrimitiveType",
//                             "value": "16"
//                         }
//                     }
//                 ],
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:3"
//                 },
//                 "FileName": "interfaces.gno",
//                 "IsMethod": false,
//                 "Name": "Render",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "interfaces.gno",
//                         "Line": "20",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "path",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "16"
//                             }
//                         }
//                     ],
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
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "IncCounter",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "11",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PrimitiveType",
//                             "value": "32"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "Counter",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "15",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "32"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "CurrentRealmPath",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "19",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
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
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "AssertOriginCall",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "23",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PrimitiveType",
//                             "value": "4"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "IsOriginCall",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "27",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "4"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "t",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests.TestRealmObject"
//                             }
//                         }
//                     }
//                 ],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "ModifyTestRealmObject",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "38",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "t",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests.TestRealmObject"
//                                 }
//                             }
//                         }
//                     ],
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "InitTestNodes",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "60",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "ModTestNodes",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "65",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "PrintTestNodes",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "73",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Realm"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "GetPrevRealm",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "77",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "std.Realm"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "std.Realm"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "GetRSubtestsPrevRealm",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "81",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "std.Realm"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "fn",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [],
//                             "Results": []
//                         }
//                     }
//                 ],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "Exec",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "85",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/demo/tests"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "fn",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [],
//                                 "Results": []
//                             }
//                         }
//                     ],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests.Stringer"
//                 },
//                 "Vrd": false
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
//                     "ID": "gno.land/r/demo/tests.TestNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests.TestNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests.TestNode"
//                 }
//             }
//         }
//     ]
// }
// switchrealm["gno.land/r/demo/tests"]
