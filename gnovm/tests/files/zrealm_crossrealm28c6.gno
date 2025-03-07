// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

type foo struct {
	name string
}

func (foo) Foo() { println("hello from crossrealm_test") }

var arr [3]*foo // arr and all elems are all real
var s []*foo

func init() {
	var f1 = foo{name: "1"}
	var f2 = foo{name: "2"}
	var f3 = foo{name: "3"}
	arr[0] = &f1
	arr[1] = &f2
	arr[2] = &f3
}

func main() {
	s = arr[:]
	println("ok")
}

// Output:
// ok

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// u[f5a516808f8976c33939133293d598ce3bca4e8d:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "IsEscaped": true,
//         "ModTime": "8",
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
//                                 "FileName": "files/zrealm_crossrealm28c6.gno",
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
//                                         "File": "files/zrealm_crossrealm28c6.gno",
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
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//                 },
//                 "Length": "3",
//                 "Maxcap": "3",
//                 "Offset": "0"
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//                 },
//                 "FileName": "files/zrealm_crossrealm28c6.gno",
//                 "IsMethod": false,
//                 "Name": "init.3",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/crossrealm_test",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "files/zrealm_crossrealm28c6.gno",
//                         "Line": "13",
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//                 },
//                 "FileName": "files/zrealm_crossrealm28c6.gno",
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
//                         "File": "files/zrealm_crossrealm28c6.gno",
//                         "Line": "22",
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
//                 "Index": "0",
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
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "IsEscaped": true,
//         "ModTime": "8",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "2"
//     }
// }
