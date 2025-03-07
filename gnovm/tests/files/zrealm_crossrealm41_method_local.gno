// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

var f func() string

type MyStruct struct{}

func (sv *MyStruct) M() string { return "a" }

var mysv *MyStruct = &MyStruct{}

func main() {
	f = mysv.M
	println(f)
}

// Output:
// <*gno.land/r/crossrealm_test.MyStruct>.M(*gno.land/r/crossrealm_test.MyStruct)(string)

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// c[f5a516808f8976c33939133293d598ce3bca4e8d:6]={
//     "Func": {
//         "Closure": {
//             "@type": "/gno.RefValue",
//             "Escaped": true,
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:5"
//         },
//         "FileName": "files/zrealm_crossrealm41_method_local.gno",
//         "IsMethod": true,
//         "Name": "M",
//         "NativeName": "",
//         "NativePkg": "",
//         "PkgPath": "gno.land/r/crossrealm_test",
//         "Source": {
//             "@type": "/gno.RefNode",
//             "BlockNode": null,
//             "Location": {
//                 "Column": "1",
//                 "File": "files/zrealm_crossrealm41_method_local.gno",
//                 "Line": "8",
//                 "PkgPath": "gno.land/r/crossrealm_test"
//             }
//         },
//         "Type": {
//             "@type": "/gno.FuncType",
//             "IsClosure": false,
//             "Params": [
//                 {
//                     "Embedded": false,
//                     "Name": "sv",
//                     "Tag": "",
//                     "Type": {
//                         "@type": "/gno.PointerType",
//                         "Elt": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/crossrealm_test.MyStruct"
//                         }
//                     }
//                 }
//             ],
//             "Results": [
//                 {
//                     "Embedded": false,
//                     "Name": "",
//                     "Tag": "",
//                     "Type": {
//                         "@type": "/gno.PrimitiveType",
//                         "value": "16"
//                     }
//                 }
//             ]
//         }
//     },
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:6",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "1"
//     },
//     "Receiver": {
//         "T": {
//             "@type": "/gno.PointerType",
//             "Elt": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/crossrealm_test.MyStruct"
//             }
//         },
//         "V": {
//             "@type": "/gno.PointerValue",
//             "Base": {
//                 "@type": "/gno.RefValue",
//                 "Escaped": true,
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//             },
//             "Index": "0",
//             "TV": null
//         }
//     }
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "IsEscaped": true,
//         "ModTime": "5",
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
//             "PkgPath": "gno.land/r/crossrealm_test"
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
//                 "@type": "/gno.RefValue",
//                 "Hash": "59f7a9616dd2e5261d77d72ccec511c627894bfa",
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:6"
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
//                         "Fields": [],
//                         "PkgPath": "gno.land/r/crossrealm_test"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "sv",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/crossrealm_test.MyStruct"
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
//                                 "FileName": "files/zrealm_crossrealm41_method_local.gno",
//                                 "IsMethod": true,
//                                 "Name": "M",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/crossrealm_test",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "files/zrealm_crossrealm41_method_local.gno",
//                                         "Line": "8",
//                                         "PkgPath": "gno.land/r/crossrealm_test"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "IsClosure": false,
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "sv",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/crossrealm_test.MyStruct"
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
//                     "Name": "MyStruct",
//                     "PkgPath": "gno.land/r/crossrealm_test"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.MyStruct"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//                 },
//                 "Index": "0",
//                 "TV": null
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:5"
//                 },
//                 "FileName": "files/zrealm_crossrealm41_method_local.gno",
//                 "IsMethod": false,
//                 "Name": "main",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/crossrealm_test",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "files/zrealm_crossrealm41_method_local.gno",
//                         "Line": "12",
//                         "PkgPath": "gno.land/r/crossrealm_test"
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
// u[f5a516808f8976c33939133293d598ce3bca4e8d:5]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:5",
//         "IsEscaped": true,
//         "ModTime": "6",
//         "RefCount": "3"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:2"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "1",
//             "File": "files/zrealm_crossrealm41_method_local.gno",
//             "Line": "2",
//             "PkgPath": "gno.land/r/crossrealm_test"
//         }
//     },
//     "Values": []
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:3]={
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "IsEscaped": true,
//         "ModTime": "6",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.MyStruct"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "c2dcd6b8b5bbe275e4cd7b90b0333aaf02e120eb",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//         }
//     }
// }
