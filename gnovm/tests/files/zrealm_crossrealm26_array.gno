// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/array"
)

var root interface{}

func main() {
	root = crossrealm.GetArray3()
	println(".")
}

// Output:
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/array"]
// switchrealm["gno.land/r/crossrealm_test"]
// c[f5a516808f8976c33939133293d598ce3bca4e8d:4]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm/array.Foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f429d463f62fe87134a149198611dfa41bd006ec:2"
//                 },
//                 "Index": "1",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "1"
//     }
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
//                 "@type": "/gno.ArrayType",
//                 "Elt": {
//                     "@type": "/gno.PointerType",
//                     "Elt": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/tests/crossrealm/array.Foo"
//                     }
//                 },
//                 "Len": "1",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "13f3494510b9d75774658a794ed0cc9189f091b8",
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
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
//                 "FileName": "files/zrealm_crossrealm26_array.gno",
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
//                         "File": "files/zrealm_crossrealm26_array.gno",
//                         "Line": "10",
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
// u[f429d463f62fe87134a149198611dfa41bd006ec:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "f429d463f62fe87134a149198611dfa41bd006ec:2",
//         "IsEscaped": true,
//         "ModTime": "4",
//         "RefCount": "3"
//     },
//     "Parent": null,
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "0",
//             "File": "",
//             "Line": "0",
//             "PkgPath": "gno.land/r/demo/tests/crossrealm/array"
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/array"
//                     },
//                     "Methods": [],
//                     "Name": "Foo",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm/array"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/array.Foo"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "fad5542d0087dcee72fea8fbb1476ee55a288180",
//                 "ObjectID": "f429d463f62fe87134a149198611dfa41bd006ec:3"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.ArrayType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm/array.Foo"
//                 },
//                 "Len": "2",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "ecaf211630f4e3180fff6746ffb474962c3dd4cc",
//                 "ObjectID": "f429d463f62fe87134a149198611dfa41bd006ec:4"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.ArrayType",
//                 "Elt": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "32"
//                 },
//                 "Len": "2",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "19bd2a83c42e0e6b152cb33b47ea30c44fa43692",
//                 "ObjectID": "f429d463f62fe87134a149198611dfa41bd006ec:7"
//             }
//         },
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
//                             "@type": "/gno.ArrayType",
//                             "Elt": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "32"
//                             },
//                             "Len": "2",
//                             "Vrd": false
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f429d463f62fe87134a149198611dfa41bd006ec:8"
//                 },
//                 "FileName": "crossrealm_array.gno",
//                 "IsMethod": false,
//                 "Name": "GetArray5",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/array",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_array.gno",
//                         "Line": "11",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/array"
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
//                                 "@type": "/gno.ArrayType",
//                                 "Elt": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "32"
//                                 },
//                                 "Len": "2",
//                                 "Vrd": false
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.ArrayType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm/array.Foo"
//                             },
//                             "Len": "2",
//                             "Vrd": false
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f429d463f62fe87134a149198611dfa41bd006ec:8"
//                 },
//                 "FileName": "crossrealm_array.gno",
//                 "IsMethod": false,
//                 "Name": "GetArray4",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/array",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_array.gno",
//                         "Line": "15",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/array"
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
//                                 "@type": "/gno.ArrayType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm/array.Foo"
//                                 },
//                                 "Len": "2",
//                                 "Vrd": false
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.ArrayType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm/array.Foo"
//                             },
//                             "Len": "2",
//                             "Vrd": false
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f429d463f62fe87134a149198611dfa41bd006ec:8"
//                 },
//                 "FileName": "crossrealm_array.gno",
//                 "IsMethod": false,
//                 "Name": "GetArray",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/array",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_array.gno",
//                         "Line": "19",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/array"
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
//                                 "@type": "/gno.ArrayType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm/array.Foo"
//                                 },
//                                 "Len": "2",
//                                 "Vrd": false
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.ArrayType",
//                             "Elt": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm/array.Foo"
//                                 }
//                             },
//                             "Len": "2",
//                             "Vrd": false
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f429d463f62fe87134a149198611dfa41bd006ec:8"
//                 },
//                 "FileName": "crossrealm_array.gno",
//                 "IsMethod": false,
//                 "Name": "GetArray2",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/array",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_array.gno",
//                         "Line": "23",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/array"
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
//                                 "@type": "/gno.ArrayType",
//                                 "Elt": {
//                                     "@type": "/gno.PointerType",
//                                     "Elt": {
//                                         "@type": "/gno.RefType",
//                                         "ID": "gno.land/r/demo/tests/crossrealm/array.Foo"
//                                     }
//                                 },
//                                 "Len": "2",
//                                 "Vrd": false
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.ArrayType",
//                             "Elt": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm/array.Foo"
//                                 }
//                             },
//                             "Len": "1",
//                             "Vrd": false
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f429d463f62fe87134a149198611dfa41bd006ec:8"
//                 },
//                 "FileName": "crossrealm_array.gno",
//                 "IsMethod": false,
//                 "Name": "GetArray3",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/array",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_array.gno",
//                         "Line": "27",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/array"
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
//                                 "@type": "/gno.ArrayType",
//                                 "Elt": {
//                                     "@type": "/gno.PointerType",
//                                     "Elt": {
//                                         "@type": "/gno.RefType",
//                                         "ID": "gno.land/r/demo/tests/crossrealm/array.Foo"
//                                     }
//                                 },
//                                 "Len": "1",
//                                 "Vrd": false
//                             }
//                         }
//                     ]
//                 }
//             }
//         }
//     ]
// }
