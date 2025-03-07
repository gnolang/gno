// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

var s []*foo

func main() {
	var arr [3]*foo

	var f1 = foo{name: "1"}
	var f2 = foo{name: "2"}
	var f3 = foo{name: "3"}
	arr[0] = &f1
	arr[1] = &f2
	arr[2] = &f3

	s = arr[1:] // will attach the whole array
	println("ok")
}

// Output:
// ok

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// c[f5a516808f8976c33939133293d598ce3bca4e8d:4]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:5"
//                 },
//                 "Index": "1",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:5"
//                 },
//                 "Index": "2",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:5"
//                 },
//                 "Index": "3",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "2"
//     }
// }
// c[f5a516808f8976c33939133293d598ce3bca4e8d:6]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "1"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:6",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:5",
//         "RefCount": "1"
//     }
// }
// c[f5a516808f8976c33939133293d598ce3bca4e8d:7]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "2"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:7",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:5",
//         "RefCount": "1"
//     }
// }
// c[f5a516808f8976c33939133293d598ce3bca4e8d:8]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "3"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:8",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:5",
//         "RefCount": "1"
//     }
// }
// c[f5a516808f8976c33939133293d598ce3bca4e8d:5]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:5",
//         "IsEscaped": true,
//         "ModTime": "0",
//         "RefCount": "3"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "1",
//             "File": "files/zrealm_crossrealm28c7.gno",
//             "Line": "12",
//             "PkgPath": "gno.land/r/crossrealm_test"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.ArrayType",
//                 "Elt": {
//                     "@type": "/gno.PointerType",
//                     "Elt": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/crossrealm_test.foo"
//                     }
//                 },
//                 "Len": "3",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Escaped": true,
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/crossrealm_test.foo"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "8c50cc7ff441ccd00157021b69d3ce46d1abfeef",
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:6"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/crossrealm_test.foo"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "099b33f8a5d45a7d9100946aebc6b7b4794f185d",
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:7"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/crossrealm_test.foo"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "d7b83413c386034a6c07d9ce9e58287d39c37998",
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:8"
//             }
//         }
//     ]
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
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
//             "PkgPath": "gno.land/r/crossrealm_test"
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
//                                 "Name": "name",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 }
//                             }
//                         ],
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
//                                         "Name": ".recv",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "gno.land/r/crossrealm_test.foo"
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "files/zrealm_crossrealm28c7.gno",
//                                 "IsMethod": true,
//                                 "Name": "Foo",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/crossrealm_test",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "files/zrealm_crossrealm28c7.gno",
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
//                                             "Name": ".recv",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/crossrealm_test.foo"
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "foo",
//                     "PkgPath": "gno.land/r/crossrealm_test"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.PointerType",
//                     "Elt": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/crossrealm_test.foo"
//                     }
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//                 },
//                 "Length": "2",
//                 "Maxcap": "2",
//                 "Offset": "1"
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//                 },
//                 "FileName": "files/zrealm_crossrealm28c7.gno",
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
//                         "File": "files/zrealm_crossrealm28c7.gno",
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
// u[f5a516808f8976c33939133293d598ce3bca4e8d:3]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "IsEscaped": true,
//         "ModTime": "8",
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
//             "File": "files/zrealm_crossrealm28c7.gno",
//             "Line": "2",
//             "PkgPath": "gno.land/r/crossrealm_test"
//         }
//     },
//     "Values": []
// }
