// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

var SS []crossrealm.XYZ

func main() {
	SS = crossrealm.GetSlice3()
	println(".")
}

// Output:
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/slice"]
// c[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:26]={
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
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:26",
//         "ModTime": "0",
//         "OwnerID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:25",
//         "RefCount": "1"
//     }
// }
// c[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:27]={
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
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:27",
//         "ModTime": "0",
//         "OwnerID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:25",
//         "RefCount": "1"
//     }
// }
// c[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:25]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "47bf5eb54da0545110d6486001609dc4203f61b0",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:26"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "eaa3017cfcfafd8f8217ee1df0afdac5e4ad12a9",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:27"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:25",
//         "ModTime": "0",
//         "OwnerID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2",
//         "RefCount": "1"
//     }
// }
// u[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2",
//         "IsEscaped": true,
//         "ModTime": "24",
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
//             "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                 "Name": "Foo",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [],
//                                     "Results": []
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     },
//                     "Methods": [],
//                     "Name": "Fooer",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.Fooer"
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
//                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.Fooer"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "SetSlice",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "9",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.Fooer"
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     },
//                     "Methods": [],
//                     "Name": "XYZ",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "0934dceb3731dd613f35a5e4a4a82396f3dc474e",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:11"
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
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "fcdb6b41091d087be521ae9be69b8328ac495677",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:25"
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
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "9073c630fa09e61b0e3b824056c022eebe7bd86b",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:16"
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
//                         "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                     }
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "088eb6ee7e7fe1c5e10101801be9e64ac6794e29",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:18"
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
//                         "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                     }
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "3450ddb4764c2026edd313f230c633e0af6cd8ed",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:4"
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
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                 },
//                 "Len": "2",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "e03c7067902bb4b3c2aa477a27d6abfcd51ce3a6",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:5"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "init.10",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "27",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "42",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "d10003fa5a9ebd5bb30cadd5e11bd5aad76ec9f4",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:8"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "8a2889ef5ccf6c7e4ec0c59ed627267fa3a51026",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:9"
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
//                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice2",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "50",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice3",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "58",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                                             "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice4",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "63",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice5",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "68",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                         "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice6",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "72",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                                             "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice7",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "77",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                 "@type": "/gno.ArrayType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
//                 },
//                 "Len": "1",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "ed9761e4c90097cbcf3d55a627ab1b7c65ff8a62",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:10"
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
//                             "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
//                         }
//                     }
//                 ],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.ArrayType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice8",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "84",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                 "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
//                             }
//                         }
//                     ],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.ArrayType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
//                                 },
//                                 "Len": "1",
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
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
//                         }
//                     }
//                 ],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.ArrayType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice9",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "89",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                 "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
//                             }
//                         }
//                     ],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.ArrayType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
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
// d[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:14]
// d[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:15]
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
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:25"
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
// u[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2",
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
//             "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                 "Name": "Foo",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [],
//                                     "Results": []
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     },
//                     "Methods": [],
//                     "Name": "Fooer",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.Fooer"
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
//                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.Fooer"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "SetSlice",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "9",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.Fooer"
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
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     },
//                     "Methods": [],
//                     "Name": "XYZ",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "0934dceb3731dd613f35a5e4a4a82396f3dc474e",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:11"
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
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:25"
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
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "9073c630fa09e61b0e3b824056c022eebe7bd86b",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:16"
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
//                         "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                     }
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "088eb6ee7e7fe1c5e10101801be9e64ac6794e29",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:18"
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
//                         "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                     }
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "3450ddb4764c2026edd313f230c633e0af6cd8ed",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:4"
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
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//                 },
//                 "Len": "2",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "e03c7067902bb4b3c2aa477a27d6abfcd51ce3a6",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:5"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "init.10",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "27",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "42",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "d10003fa5a9ebd5bb30cadd5e11bd5aad76ec9f4",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:8"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "8a2889ef5ccf6c7e4ec0c59ed627267fa3a51026",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:9"
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
//                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice2",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "50",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice3",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "58",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                                             "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice4",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "63",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice5",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "68",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                         "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice6",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "72",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                                             "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice7",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "77",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
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
//                 "@type": "/gno.ArrayType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
//                 },
//                 "Len": "1",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "ed9761e4c90097cbcf3d55a627ab1b7c65ff8a62",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:10"
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
//                             "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
//                         }
//                     }
//                 ],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.ArrayType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice8",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "84",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                 "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
//                             }
//                         }
//                     ],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.ArrayType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
//                                 },
//                                 "Len": "1",
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
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
//                         }
//                     }
//                 ],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.ArrayType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "GetSlice9",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "89",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                 "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
//                             }
//                         }
//                     ],
//                     "Results": [
//                         {
//                             "Embedded": false,
//                             "Name": "",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.ArrayType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/p/demo/tests/p_crossrealm.Stringer"
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
