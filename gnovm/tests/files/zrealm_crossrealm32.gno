// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
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
// c[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:30]={
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
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:30",
//         "ModTime": "0",
//         "OwnerID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "RefCount": "1"
//     }
// }
// c[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:31]={
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
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:31",
//         "ModTime": "0",
//         "OwnerID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "RefCount": "1"
//     }
// }
// c[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "de4d4b9346995296fbd8210badf8bab1c111e60c",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:30"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "1570e95bdb3313b5a0656b56b9bba96c9c7704aa",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:31"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
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
//         "ModTime": "28",
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
//                                     "IsClosure": false,
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
//                 "IsClosure": false,
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
//                     "IsClosure": false,
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
//                 "@type": "/gno.ArrayType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.Fooer"
//                 },
//                 "Len": "3",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "d62efae4bccfb771ca26b5401811c88b7a4d4600",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:4"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "arr",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.ArrayType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.Fooer"
//                             },
//                             "Len": "3",
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
//                 "Name": "SetArray",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "18",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "arr",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.ArrayType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.Fooer"
//                                 },
//                                 "Len": "3",
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
//                     "Hash": "c6af76356d551443236418387570bc8842e2f13d",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:15"
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
//                     "Hash": "ead7f89f5eecedb66a0d1c9018643b765c445606",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29"
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
//                     "Hash": "48a643765d27caf33f1a84e6cd2626c09de575ec",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:5"
//                 },
//                 "Length": "1",
//                 "Maxcap": "2",
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
//                     "Hash": "4a9542d3dbc515401ea899e520e20985a29793eb",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:21"
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
//                     "Hash": "efa17cc1bdfcc78d6ce7fd3d37d8eb1463d2f3f8",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:8"
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
//                 "Hash": "9128ce5efd0b78fe0b2508620d33297efc8ca400",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:9"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "32"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "3d8f45f94de2ae064cb1b22cb23131014b5e2b3e",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:28"
//                 },
//                 "Length": "1",
//                 "Maxcap": "1",
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "init.13",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "43",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                         "Line": "60",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                 "Hash": "4e519b2e696d045f540a772835d23f935a34a418",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:12"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "767f878ff4789535ef68989918de26fe5a792ee0",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:13"
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
//                         "Line": "68",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                 "IsClosure": false,
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
//                         "Line": "76",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "IsClosure": false,
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
//                         "Line": "81",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
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
//                 "IsClosure": false,
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
//                         "Line": "86",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "IsClosure": false,
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
//                         "Line": "90",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
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
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "IsClosure": false,
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
//                         "Line": "95",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
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
//                 "Hash": "dc790554954efbbee8c7251652faacd0ffe4f588",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:14"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "IsClosure": false,
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
//                         "Line": "102",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
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
//                 "IsClosure": false,
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
//                         "Line": "107",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
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
//                 "IsClosure": false,
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.SliceType",
//                             "Elt": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "32"
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
//                 "Name": "GetSlice10",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "114",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                 "@type": "/gno.SliceType",
//                                 "Elt": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "32"
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
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "IsClosure": false,
//                             "Params": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "s",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.SliceType",
//                                         "Elt": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "32"
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
//                 "Name": "GetSlice11",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "118",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "s",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.SliceType",
//                                             "Elt": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "32"
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
//         }
//     ]
// }
// d[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:18]
// d[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:19]
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29"
//                 },
//                 "Length": "2",
//                 "Maxcap": "2",
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//                 },
//                 "FileName": "files/zrealm_crossrealm32.gno",
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
//                         "File": "files/zrealm_crossrealm32.gno",
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
// u[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "de4d4b9346995296fbd8210badf8bab1c111e60c",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:30"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "1570e95bdb3313b5a0656b56b9bba96c9c7704aa",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:31"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "RefCount": "2"
//     }
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
//                                     "IsClosure": false,
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
//                 "IsClosure": false,
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
//                     "IsClosure": false,
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
//                 "@type": "/gno.ArrayType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm/slice.Fooer"
//                 },
//                 "Len": "3",
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "d62efae4bccfb771ca26b5401811c88b7a4d4600",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:4"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "arr",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.ArrayType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm/slice.Fooer"
//                             },
//                             "Len": "3",
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
//                 "Name": "SetArray",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "18",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "arr",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.ArrayType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm/slice.Fooer"
//                                 },
//                                 "Len": "3",
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
//                     "Hash": "c6af76356d551443236418387570bc8842e2f13d",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:15"
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29"
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
//                     "Hash": "48a643765d27caf33f1a84e6cd2626c09de575ec",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:5"
//                 },
//                 "Length": "1",
//                 "Maxcap": "2",
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
//                     "Hash": "4a9542d3dbc515401ea899e520e20985a29793eb",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:21"
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
//                     "Hash": "efa17cc1bdfcc78d6ce7fd3d37d8eb1463d2f3f8",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:8"
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
//                 "Hash": "9128ce5efd0b78fe0b2508620d33297efc8ca400",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:9"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.PrimitiveType",
//                     "value": "32"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "3d8f45f94de2ae064cb1b22cb23131014b5e2b3e",
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:28"
//                 },
//                 "Length": "1",
//                 "Maxcap": "1",
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
//                     "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:3"
//                 },
//                 "FileName": "crossrealm_slice.gno",
//                 "IsMethod": false,
//                 "Name": "init.13",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "43",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                         "Line": "60",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                 "Hash": "4e519b2e696d045f540a772835d23f935a34a418",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:12"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "767f878ff4789535ef68989918de26fe5a792ee0",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:13"
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
//                         "Line": "68",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                 "IsClosure": false,
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
//                         "Line": "76",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "IsClosure": false,
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
//                         "Line": "81",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
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
//                 "IsClosure": false,
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
//                         "Line": "86",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "IsClosure": false,
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
//                         "Line": "90",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
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
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "IsClosure": false,
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
//                         "Line": "95",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
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
//                 "Hash": "dc790554954efbbee8c7251652faacd0ffe4f588",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:14"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "IsClosure": false,
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
//                         "Line": "102",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
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
//                 "IsClosure": false,
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
//                         "Line": "107",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
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
//                 "IsClosure": false,
//                 "Params": [],
//                 "Results": [
//                     {
//                         "Embedded": false,
//                         "Name": "",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.SliceType",
//                             "Elt": {
//                                 "@type": "/gno.PrimitiveType",
//                                 "value": "32"
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
//                 "Name": "GetSlice10",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "114",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
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
//                                 "@type": "/gno.SliceType",
//                                 "Elt": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "32"
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
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "f",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.FuncType",
//                             "IsClosure": false,
//                             "Params": [
//                                 {
//                                     "Embedded": false,
//                                     "Name": "s",
//                                     "Tag": "",
//                                     "Type": {
//                                         "@type": "/gno.SliceType",
//                                         "Elt": {
//                                             "@type": "/gno.PrimitiveType",
//                                             "value": "32"
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
//                 "Name": "GetSlice11",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm/slice",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_slice.gno",
//                         "Line": "118",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm/slice"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "f",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "s",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.SliceType",
//                                             "Elt": {
//                                                 "@type": "/gno.PrimitiveType",
//                                                 "value": "32"
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
//         }
//     ]
// }
