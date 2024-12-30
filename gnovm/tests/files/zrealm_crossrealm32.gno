// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm"
)

var SS []crossrealm.XYZ

func main() {
	SS = crossrealm.GetSlice3()
	println(".")
}

// Output:
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:46]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "0"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:46",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:45",
//         "RefCount": "1"
//     }
// }
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:45]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Escaped": true,
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "2489b816b7c9cfcc07f7e853718aaa019016fced",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:46"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:45",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//         "RefCount": "1"
//     }
// }
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//         "IsEscaped": true,
//         "ModTime": "44",
//         "RefCount": "10"
//     },
//     "Parent": null,
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "0",
//             "File": "",
//             "Line": "0",
//             "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.F"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "ff",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests/crossrealm.F"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:3"
//                 },
//                 "FileName": "crossrealm_func.gno",
//                 "IsMethod": false,
//                 "Name": "SetCallback",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func.gno",
//                         "Line": "7",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "ff",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.F"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:3"
//                 },
//                 "FileName": "crossrealm_func.gno",
//                 "IsMethod": false,
//                 "Name": "ExecuteCallback",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func.gno",
//                         "Line": "11",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:4"
//                 },
//                 "FileName": "",
//                 "IsMethod": false,
//                 "Name": "",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "10",
//                         "File": "crossrealm_func2.gno",
//                         "Line": "3",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:4"
//                 },
//                 "FileName": "crossrealm_func2.gno",
//                 "IsMethod": false,
//                 "Name": "GetFunc",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func2.gno",
//                         "Line": "8",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.FuncType",
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
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:4"
//                 },
//                 "FileName": "crossrealm_func2.gno",
//                 "IsMethod": false,
//                 "Name": "GetFunc2",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func2.gno",
//                         "Line": "12",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.FuncType",
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "sv",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.MyStruct"
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
//                                 "FileName": "crossrealm_func2.gno",
//                                 "IsMethod": true,
//                                 "Name": "M",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "crossrealm_func2.gno",
//                                         "Line": "20",
//                                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "sv",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/demo/tests/crossrealm.MyStruct"
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
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.MyStruct"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "66606f1f6d50797bbe89640247170078dcbf0e6a",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:5"
//                 },
//                 "Index": "0",
//                 "TV": null
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
//                             "@type": "/gno.FuncType",
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:4"
//                 },
//                 "FileName": "crossrealm_func2.gno",
//                 "IsMethod": false,
//                 "Name": "GetMethod",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func2.gno",
//                         "Line": "30",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.FuncType",
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
//                         "@type": "/gno.InterfaceType",
//                         "Generic": "",
//                         "Methods": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "Foo",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [],
//                                     "Results": []
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "Fooer",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
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
//                             "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:7"
//                 },
//                 "FileName": "crossrealm_iface.gno",
//                 "IsMethod": false,
//                 "Name": "SetFooer",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_iface.gno",
//                         "Line": "8",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
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
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:7"
//                 },
//                 "FileName": "crossrealm_iface.gno",
//                 "IsMethod": false,
//                 "Name": "CallFoo",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_iface.gno",
//                         "Line": "13",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "f",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "Container",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.Container"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "a7562372e10ade1d106ae5ef9e97c170ee4b50f1",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:8"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:7"
//                 },
//                 "FileName": "crossrealm_iface.gno",
//                 "IsMethod": false,
//                 "Name": "SetContainer",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_iface.gno",
//                         "Line": "23",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
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
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:7"
//                 },
//                 "FileName": "crossrealm_iface.gno",
//                 "IsMethod": false,
//                 "Name": "SetContainer2",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_iface.gno",
//                         "Line": "27",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                             }
//                         }
//                     ],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.MapType",
//                 "Key": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "16"
//                 },
//                 "Value": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "32"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "cda4bc373a1cceedd07fe9fd29410bfd5732df7c",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:25"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:9"
//                 },
//                 "FileName": "crossrealm_map.gno",
//                 "IsMethod": false,
//                 "Name": "init.19",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_map.gno",
//                         "Line": "5",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                             "@type": "/gno.MapType",
//                             "Key": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "16"
//                             },
//                             "Value": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "32"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:9"
//                 },
//                 "FileName": "crossrealm_map.gno",
//                 "IsMethod": false,
//                 "Name": "GetMap",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_map.gno",
//                         "Line": "10",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.MapType",
//                                 "Key": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 },
//                                 "Value": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "32"
//                                 }
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
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "A",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "32"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "ls",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.LocalStruct"
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
//                                 "FileName": "crossrealm_p.gno",
//                                 "IsMethod": true,
//                                 "Name": "String",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "crossrealm_p.gno",
//                                         "Line": "12",
//                                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "ls",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/demo/tests/crossrealm.LocalStruct"
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
//                     "Name": "LocalStruct",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.LocalStruct"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "921bae5520c2c8b9d456b47651ea80481836de4e",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:26"
//                 },
//                 "Index": "0",
//                 "TV": null
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:10"
//                 },
//                 "FileName": "crossrealm_p.gno",
//                 "IsMethod": false,
//                 "Name": "init.23",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_p.gno",
//                         "Line": "19",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/p/demo/tests/p_crossrealm.Container"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:10"
//                 },
//                 "FileName": "crossrealm_p.gno",
//                 "IsMethod": false,
//                 "Name": "Make1",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_p.gno",
//                         "Line": "24",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                     "ID": "gno.land/p/demo/tests/p_crossrealm.Container"
//                                 }
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                 },
//                 "Vrd": false
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "fs",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.SliceType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                             },
//                             "Vrd": false
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:11"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "SetSlice",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "5",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "fs",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.SliceType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                                 },
//                                 "Vrd": false
//                             }
//                         }
//                     ],
//                     "Results": []
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "XYZ",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "b3b42c06d65da2199792616887b55d65b79f075e",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:28"
//                 },
//                 "Length": "2",
//                 "Maxcap": "2",
//                 "Offset": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "fa5658dcf95bafce7154d3380ebf7e9211b4ab4d",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:45"
//                 },
//                 "Length": "2",
//                 "Maxcap": "2",
//                 "Offset": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "7c8932f1337c257a3236c79d7d8b4211812d8385",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:33"
//                 },
//                 "Length": "1",
//                 "Maxcap": "1",
//                 "Offset": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.PointerType",
//                     "Elt": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                     }
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "dbf56198a94f3a1d98c501d1b8c58b88bfbe8bce",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:35"
//                 },
//                 "Length": "1",
//                 "Maxcap": "1",
//                 "Offset": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.PointerType",
//                     "Elt": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                     }
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "8f5471d881fd68800170d03c52907ce7a7607d09",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:12"
//                 },
//                 "Length": "2",
//                 "Maxcap": "2",
//                 "Offset": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.ArrayType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                 },
//                 "Len": "2",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "41d4fd3b5431d4cd8494803a01f437bddd45b521",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:13"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "init.34",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "13",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                             "@type": "/gno.SliceType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                             },
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "28",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.SliceType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                 },
//                                 "Vrd": false
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "b439d9527c852f481b3dcc9c4ad27fcf6b2cbc6e",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:17"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "c3a08360b77971b3c8a35a4c9c0955f6336760c4",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:18"
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
//                             "@type": "/gno.SliceType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                             },
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice2",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "36",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.SliceType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                 },
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
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.SliceType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                             },
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice3",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "44",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.SliceType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                 },
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
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "s",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.SliceType",
//                                         "Elt": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                         },
//                                         "Vrd": false
//                                     }
//                                 }
//                             ],
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice4",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "49",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "s",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.SliceType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                             },
//                                             "Vrd": false
//                                         }
//                                     }
//                                 ],
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
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.SliceType",
//                             "Elt": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                 }
//                             },
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice5",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "54",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.SliceType",
//                                 "Elt": {
//                                     "@type": "/gno.PointerType",
//                                     "Elt": {
//                                         "@type": "/gno.RefType",
//                                         "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                     }
//                                 },
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
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "s",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.SliceType",
//                                         "Elt": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                             }
//                                         },
//                                         "Vrd": false
//                                     }
//                                 }
//                             ],
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice6",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "58",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "s",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.SliceType",
//                                             "Elt": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                                 }
//                                             },
//                                             "Vrd": false
//                                         }
//                                     }
//                                 ],
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
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "s",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.ArrayType",
//                                         "Elt": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                         },
//                                         "Len": "2",
//                                         "Vrd": false
//                                     }
//                                 }
//                             ],
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice7",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "63",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "s",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.ArrayType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                             },
//                                             "Len": "2",
//                                             "Vrd": false
//                                         }
//                                     }
//                                 ],
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
//                                 "Name": "A",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "32"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "b",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                                             }
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "crossrealm_struct.gno",
//                                 "IsMethod": true,
//                                 "Name": "String",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "crossrealm_struct.gno",
//                                         "Line": "12",
//                                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "b",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                                                 }
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
//                                         "Name": "b",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                                             }
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "crossrealm_struct.gno",
//                                 "IsMethod": true,
//                                 "Name": "ModifyBar",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "crossrealm_struct.gno",
//                                         "Line": "26",
//                                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "b",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                                                 }
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "Bar",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "eb498518bc671b8c1625ffbe54846ac1df8601a5",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:19"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "b",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                             }
//                         }
//                     }
//                 ],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Bar"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:21"
//                 },
//                 "FileName": "crossrealm_struct.gno",
//                 "IsMethod": false,
//                 "Name": "SetBar",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_struct.gno",
//                         "Line": "16",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "b",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                                 }
//                             }
//                         }
//                     ],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                                 }
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:21"
//                 },
//                 "FileName": "crossrealm_struct.gno",
//                 "IsMethod": false,
//                 "Name": "ChangeBar",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_struct.gno",
//                         "Line": "21",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "A",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "B",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "D",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "Embedded": true,
//                                 "Name": "A",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.A"
//                                 }
//                             },
//                             {
//                                 "Embedded": true,
//                                 "Name": "B",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.B"
//                                 }
//                             },
//                             {
//                                 "Embedded": true,
//                                 "Name": "D",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.D"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "C",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.A"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "600bc777de2ebe4f91378028f2331fdfdc891ddc",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:22"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.B"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "78dd21cd9cef4925a96ee4fbbf182d1494f646d5",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:23"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.C"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "c443db3ae26b6a807ae6a997f395af8ca986b3b3",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:42"
//                 },
//                 "Index": "0",
//                 "TV": null
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:24"
//                 },
//                 "FileName": "crossrealm_struct2.gno",
//                 "IsMethod": false,
//                 "Name": "init.56",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_struct2.gno",
//                         "Line": "26",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "cb",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "v",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.PointerType",
//                                         "Elt": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "gno.land/r/demo/tests/crossrealm.C"
//                                         }
//                                     }
//                                 }
//                             ],
//                             "Results": []
//                         }
//                     }
//                 ],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.C"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:24"
//                 },
//                 "FileName": "crossrealm_struct2.gno",
//                 "IsMethod": false,
//                 "Name": "GetStruct",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_struct2.gno",
//                         "Line": "32",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "cb",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "v",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.C"
//                                             }
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             }
//                         }
//                     ],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.C"
//                                 }
//                             }
//                         }
//                     ]
//                 }
//             }
//         }
//     ]
// }
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:31]
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:32]
// switchrealm["gno.land/r/crossrealm_test"]
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
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:45"
//                 },
//                 "Length": "2",
//                 "Maxcap": "2",
//                 "Offset": "0"
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//                 },
//                 "FileName": "main.gno",
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
//                         "File": "main.gno",
//                         "Line": "12",
//                         "PkgPath": "gno.land/r/crossrealm_test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": []
//                 }
//             }
//         }
//     ]
// }
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "RefCount": "10"
//     },
//     "Parent": null,
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "0",
//             "File": "",
//             "Line": "0",
//             "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.F"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "ff",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests/crossrealm.F"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:3"
//                 },
//                 "FileName": "crossrealm_func.gno",
//                 "IsMethod": false,
//                 "Name": "SetCallback",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func.gno",
//                         "Line": "7",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "ff",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.F"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:3"
//                 },
//                 "FileName": "crossrealm_func.gno",
//                 "IsMethod": false,
//                 "Name": "ExecuteCallback",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func.gno",
//                         "Line": "11",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:4"
//                 },
//                 "FileName": "",
//                 "IsMethod": false,
//                 "Name": "",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "10",
//                         "File": "crossrealm_func2.gno",
//                         "Line": "3",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:4"
//                 },
//                 "FileName": "crossrealm_func2.gno",
//                 "IsMethod": false,
//                 "Name": "GetFunc",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func2.gno",
//                         "Line": "8",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.FuncType",
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
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:4"
//                 },
//                 "FileName": "crossrealm_func2.gno",
//                 "IsMethod": false,
//                 "Name": "GetFunc2",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func2.gno",
//                         "Line": "12",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.FuncType",
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "sv",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.MyStruct"
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
//                                 "FileName": "crossrealm_func2.gno",
//                                 "IsMethod": true,
//                                 "Name": "M",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "crossrealm_func2.gno",
//                                         "Line": "20",
//                                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "sv",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/demo/tests/crossrealm.MyStruct"
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
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.MyStruct"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "66606f1f6d50797bbe89640247170078dcbf0e6a",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:5"
//                 },
//                 "Index": "0",
//                 "TV": null
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
//                             "@type": "/gno.FuncType",
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:4"
//                 },
//                 "FileName": "crossrealm_func2.gno",
//                 "IsMethod": false,
//                 "Name": "GetMethod",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_func2.gno",
//                         "Line": "30",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.FuncType",
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
//                         "@type": "/gno.InterfaceType",
//                         "Generic": "",
//                         "Methods": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "Foo",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [],
//                                     "Results": []
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "Fooer",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
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
//                             "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:7"
//                 },
//                 "FileName": "crossrealm_iface.gno",
//                 "IsMethod": false,
//                 "Name": "SetFooer",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_iface.gno",
//                         "Line": "8",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
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
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:7"
//                 },
//                 "FileName": "crossrealm_iface.gno",
//                 "IsMethod": false,
//                 "Name": "CallFoo",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_iface.gno",
//                         "Line": "13",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "f",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "Container",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.Container"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "a7562372e10ade1d106ae5ef9e97c170ee4b50f1",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:8"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:7"
//                 },
//                 "FileName": "crossrealm_iface.gno",
//                 "IsMethod": false,
//                 "Name": "SetContainer",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_iface.gno",
//                         "Line": "23",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
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
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:7"
//                 },
//                 "FileName": "crossrealm_iface.gno",
//                 "IsMethod": false,
//                 "Name": "SetContainer2",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_iface.gno",
//                         "Line": "27",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                             }
//                         }
//                     ],
//                     "Results": []
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.MapType",
//                 "Key": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "16"
//                 },
//                 "Value": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "32"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "cda4bc373a1cceedd07fe9fd29410bfd5732df7c",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:25"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:9"
//                 },
//                 "FileName": "crossrealm_map.gno",
//                 "IsMethod": false,
//                 "Name": "init.19",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_map.gno",
//                         "Line": "5",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                             "@type": "/gno.MapType",
//                             "Key": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "16"
//                             },
//                             "Value": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "32"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:9"
//                 },
//                 "FileName": "crossrealm_map.gno",
//                 "IsMethod": false,
//                 "Name": "GetMap",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_map.gno",
//                         "Line": "10",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.MapType",
//                                 "Key": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "16"
//                                 },
//                                 "Value": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "32"
//                                 }
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
//                         "Fields": [
//                             {
//                                 "Embedded": false,
//                                 "Name": "A",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "32"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "ls",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.LocalStruct"
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
//                                 "FileName": "crossrealm_p.gno",
//                                 "IsMethod": true,
//                                 "Name": "String",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "crossrealm_p.gno",
//                                         "Line": "12",
//                                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "ls",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/demo/tests/crossrealm.LocalStruct"
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
//                     "Name": "LocalStruct",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.LocalStruct"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "921bae5520c2c8b9d456b47651ea80481836de4e",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:26"
//                 },
//                 "Index": "0",
//                 "TV": null
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:10"
//                 },
//                 "FileName": "crossrealm_p.gno",
//                 "IsMethod": false,
//                 "Name": "init.23",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_p.gno",
//                         "Line": "19",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/p/demo/tests/p_crossrealm.Container"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:10"
//                 },
//                 "FileName": "crossrealm_p.gno",
//                 "IsMethod": false,
//                 "Name": "Make1",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_p.gno",
//                         "Line": "24",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                     "ID": "gno.land/p/demo/tests/p_crossrealm.Container"
//                                 }
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                 },
//                 "Vrd": false
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "fs",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.SliceType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                             },
//                             "Vrd": false
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:11"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "SetSlice",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "5",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "fs",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.SliceType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                                 },
//                                 "Vrd": false
//                             }
//                         }
//                     ],
//                     "Results": []
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "XYZ",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "b3b42c06d65da2199792616887b55d65b79f075e",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:28"
//                 },
//                 "Length": "2",
//                 "Maxcap": "2",
//                 "Offset": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:45"
//                 },
//                 "Length": "2",
//                 "Maxcap": "2",
//                 "Offset": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "7c8932f1337c257a3236c79d7d8b4211812d8385",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:33"
//                 },
//                 "Length": "1",
//                 "Maxcap": "1",
//                 "Offset": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.PointerType",
//                     "Elt": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                     }
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "dbf56198a94f3a1d98c501d1b8c58b88bfbe8bce",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:35"
//                 },
//                 "Length": "1",
//                 "Maxcap": "1",
//                 "Offset": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.PointerType",
//                     "Elt": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                     }
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "8f5471d881fd68800170d03c52907ce7a7607d09",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:12"
//                 },
//                 "Length": "2",
//                 "Maxcap": "2",
//                 "Offset": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.ArrayType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                 },
//                 "Len": "2",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "41d4fd3b5431d4cd8494803a01f437bddd45b521",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:13"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "init.34",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "13",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                             "@type": "/gno.SliceType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                             },
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "28",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.SliceType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                 },
//                                 "Vrd": false
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "b439d9527c852f481b3dcc9c4ad27fcf6b2cbc6e",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:17"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "c3a08360b77971b3c8a35a4c9c0955f6336760c4",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:18"
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
//                             "@type": "/gno.SliceType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                             },
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice2",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "36",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.SliceType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                 },
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
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.SliceType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                             },
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice3",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "44",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.SliceType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                 },
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
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "s",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.SliceType",
//                                         "Elt": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                         },
//                                         "Vrd": false
//                                     }
//                                 }
//                             ],
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice4",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "49",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "s",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.SliceType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                             },
//                                             "Vrd": false
//                                         }
//                                     }
//                                 ],
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
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.SliceType",
//                             "Elt": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                 }
//                             },
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice5",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "54",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.SliceType",
//                                 "Elt": {
//                                     "@type": "/gno.PointerType",
//                                     "Elt": {
//                                         "@type": "/gno.RefType",
//                                         "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                     }
//                                 },
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
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "s",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.SliceType",
//                                         "Elt": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                             }
//                                         },
//                                         "Vrd": false
//                                     }
//                                 }
//                             ],
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice6",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "58",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "s",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.SliceType",
//                                             "Elt": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                                 }
//                                             },
//                                             "Vrd": false
//                                         }
//                                     }
//                                 ],
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
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "s",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.ArrayType",
//                                         "Elt": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                         },
//                                         "Len": "2",
//                                         "Vrd": false
//                                     }
//                                 }
//                             ],
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:16"
//                 },
//                 "FileName": "crossrealm_slice2.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice7",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice2.gno",
//                         "Line": "63",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "s",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.ArrayType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.XYZ"
//                                             },
//                                             "Len": "2",
//                                             "Vrd": false
//                                         }
//                                     }
//                                 ],
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
//                                 "Name": "A",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "32"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "b",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                                             }
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "crossrealm_struct.gno",
//                                 "IsMethod": true,
//                                 "Name": "String",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "crossrealm_struct.gno",
//                                         "Line": "12",
//                                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "b",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                                                 }
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
//                                         "Name": "b",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                                             }
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "crossrealm_struct.gno",
//                                 "IsMethod": true,
//                                 "Name": "ModifyBar",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "crossrealm_struct.gno",
//                                         "Line": "26",
//                                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "b",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.PointerType",
//                                                 "Elt": {
//                                                     "@type": "/gno.RefType",
//                                                     "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                                                 }
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "Bar",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "eb498518bc671b8c1625ffbe54846ac1df8601a5",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:19"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "b",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                             }
//                         }
//                     }
//                 ],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.Bar"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:21"
//                 },
//                 "FileName": "crossrealm_struct.gno",
//                 "IsMethod": false,
//                 "Name": "SetBar",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_struct.gno",
//                         "Line": "16",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "b",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                                 }
//                             }
//                         }
//                     ],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.Bar"
//                                 }
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:21"
//                 },
//                 "FileName": "crossrealm_struct.gno",
//                 "IsMethod": false,
//                 "Name": "ChangeBar",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_struct.gno",
//                         "Line": "21",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "A",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "B",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "D",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "Embedded": true,
//                                 "Name": "A",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.A"
//                                 }
//                             },
//                             {
//                                 "Embedded": true,
//                                 "Name": "B",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.B"
//                                 }
//                             },
//                             {
//                                 "Embedded": true,
//                                 "Name": "D",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.D"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     },
//                     "Methods": [],
//                     "Name": "C",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.A"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "600bc777de2ebe4f91378028f2331fdfdc891ddc",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:22"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.B"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "78dd21cd9cef4925a96ee4fbbf182d1494f646d5",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:23"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.C"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "c443db3ae26b6a807ae6a997f395af8ca986b3b3",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:42"
//                 },
//                 "Index": "0",
//                 "TV": null
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:24"
//                 },
//                 "FileName": "crossrealm_struct2.gno",
//                 "IsMethod": false,
//                 "Name": "init.56",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_struct2.gno",
//                         "Line": "26",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "cb",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "Params": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "v",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.PointerType",
//                                         "Elt": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "gno.land/r/demo/tests/crossrealm.C"
//                                         }
//                                     }
//                                 }
//                             ],
//                             "Results": []
//                         }
//                     }
//                 ],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.C"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:24"
//                 },
//                 "FileName": "crossrealm_struct2.gno",
//                 "IsMethod": false,
//                 "Name": "GetStruct",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_struct2.gno",
//                         "Line": "32",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "cb",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "v",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests/crossrealm.C"
//                                             }
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             }
//                         }
//                     ],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.C"
//                                 }
//                             }
//                         }
//                     ]
//                 }
//             }
//         }
//     ]
// }
