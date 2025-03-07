// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/func"
)

var f = func() bool {
	println("callback")
	return true
}

func main() {
	crossrealm.SetCallback(f)
	crossrealm.ExecuteCallback()
	println("ok")
}

// Output:
// callback
// ok

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/func"]
// u[8a0b7b87d40c36f8e8c7f01ec200a2df0f884738:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "8a0b7b87d40c36f8e8c7f01ec200a2df0f884738:2",
//         "IsEscaped": true,
//         "ModTime": "6",
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
//             "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
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
//                         "@type": "/gno.FuncType",
//                         "IsClosure": false,
//                         "Params": [],
//                         "Results": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "4"
//                                 }
//                             }
//                         ]
//                     },
//                     "Methods": [],
//                     "Name": "F",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/func.F"
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//                 },
//                 "FileName": "",
//                 "IsMethod": false,
//                 "Name": "",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/crossrealm_test",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "9",
//                         "File": "files/zrealm_crossrealm22d.gno",
//                         "Line": "8",
//                         "PkgPath": "gno.land/r/crossrealm_test"
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
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "ff",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests/crossrealm/func.F"
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
//                     "ObjectID": "8a0b7b87d40c36f8e8c7f01ec200a2df0f884738:3"
//                 },
//                 "FileName": "crossrealm_func.gno",
//                 "IsMethod": false,
//                 "Name": "SetCallback",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/func",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func.gno",
//                         "Line": "7",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "ff",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm/func.F"
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
//                 "IsClosure": false,
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "8a0b7b87d40c36f8e8c7f01ec200a2df0f884738:3"
//                 },
//                 "FileName": "crossrealm_func.gno",
//                 "IsMethod": false,
//                 "Name": "ExecuteCallback",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/func",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func.gno",
//                         "Line": "11",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
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
//                     "ObjectID": "8a0b7b87d40c36f8e8c7f01ec200a2df0f884738:3"
//                 },
//                 "FileName": "",
//                 "IsMethod": false,
//                 "Name": "",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/func",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "10",
//                         "File": "crossrealm_func.gno",
//                         "Line": "15",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "IsClosure": false,
//                             "Params": [],
//                             "Results": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.PrimitiveType",
//                                         "value": "16"
//                                     }
//                                 }
//                             ]
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "8a0b7b87d40c36f8e8c7f01ec200a2df0f884738:3"
//                 },
//                 "FileName": "crossrealm_func.gno",
//                 "IsMethod": false,
//                 "Name": "GetFunc",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/func",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func.gno",
//                         "Line": "20",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
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
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
//                                 "Params": [],
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
//                             "@type": "/gno.FuncType",
//                             "IsClosure": false,
//                             "Params": [],
//                             "Results": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.PrimitiveType",
//                                         "value": "16"
//                                     }
//                                 }
//                             ]
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "8a0b7b87d40c36f8e8c7f01ec200a2df0f884738:3"
//                 },
//                 "FileName": "crossrealm_func.gno",
//                 "IsMethod": false,
//                 "Name": "GetFunc2",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/func",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func.gno",
//                         "Line": "24",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
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
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
//                                 "Params": [],
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
//                             }
//                         }
//                     ]
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
//                         "Fields": [],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
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
//                                                 "ID": "gno.land/r/demo/tests/crossrealm/func.MyStruct"
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
//                                 "FileName": "crossrealm_func.gno",
//                                 "IsMethod": true,
//                                 "Name": "M",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/demo/tests/crossrealm/func",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "crossrealm_func.gno",
//                                         "Line": "33",
//                                         "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
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
//                                                     "ID": "gno.land/r/demo/tests/crossrealm/func.MyStruct"
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
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
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
//                                 "Name": "name",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "sv2",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "gno.land/r/demo/tests/crossrealm/func.MyStruct2"
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
//                                 "FileName": "crossrealm_func.gno",
//                                 "IsMethod": true,
//                                 "Name": "M2",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/demo/tests/crossrealm/func",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "crossrealm_func.gno",
//                                         "Line": "37",
//                                         "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "IsClosure": false,
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "sv2",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm/func.MyStruct2"
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
//                     "Name": "MyStruct2",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm/func.MyStruct"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "760805e79a860437ff34758d496a55dc66102cca",
//                     "ObjectID": "8a0b7b87d40c36f8e8c7f01ec200a2df0f884738:4"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/func.MyStruct2"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "afe5b4e878dc43b66ef9c46cd3bd0edab29c7a15",
//                 "ObjectID": "8a0b7b87d40c36f8e8c7f01ec200a2df0f884738:6"
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
//                             "@type": "/gno.FuncType",
//                             "IsClosure": false,
//                             "Params": [],
//                             "Results": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.PrimitiveType",
//                                         "value": "16"
//                                     }
//                                 }
//                             ]
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "8a0b7b87d40c36f8e8c7f01ec200a2df0f884738:3"
//                 },
//                 "FileName": "crossrealm_func.gno",
//                 "IsMethod": false,
//                 "Name": "GetMethod",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/func",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func.gno",
//                         "Line": "41",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
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
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
//                                 "Params": [],
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
//                             "@type": "/gno.FuncType",
//                             "IsClosure": false,
//                             "Params": [],
//                             "Results": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.PrimitiveType",
//                                         "value": "16"
//                                     }
//                                 }
//                             ]
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "8a0b7b87d40c36f8e8c7f01ec200a2df0f884738:3"
//                 },
//                 "FileName": "crossrealm_func.gno",
//                 "IsMethod": false,
//                 "Name": "GetMethod2",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/func",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func.gno",
//                         "Line": "45",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
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
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
//                                 "Params": [],
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
//                             }
//                         }
//                     ]
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
//         "ModTime": "6",
//         "RefCount": "4"
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
//             "File": "files/zrealm_crossrealm22d.gno",
//             "Line": "2",
//             "PkgPath": "gno.land/r/crossrealm_test"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.PackageType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/func"
//             }
//         }
//     ]
// }
// switchrealm["gno.land/r/crossrealm_test"]
// switchrealm["gno.land/r/demo/tests/crossrealm/func"]
// switchrealm["gno.land/r/crossrealm_test"]
