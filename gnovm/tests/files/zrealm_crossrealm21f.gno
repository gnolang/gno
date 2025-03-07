// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/struct"

var s *crossrealm.C

func cb(c *crossrealm.C) {
	s = c
}
func main() {
	crossrealm.GetStruct(cb)
	println(s)
}

// Output:
// &(struct{(ref(5965ae851c02ab677bc8394b408535c1db9b2635:11) gno.land/r/demo/tests/crossrealm/struct.A),(ref(5965ae851c02ab677bc8394b408535c1db9b2635:12) gno.land/r/demo/tests/crossrealm/struct.B),(struct{("d" string)} gno.land/r/demo/tests/crossrealm/struct.D)} gno.land/r/demo/tests/crossrealm/struct.C)

// Realm:
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
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm/struct.C"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "5965ae851c02ab677bc8394b408535c1db9b2635:9"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "c",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm/struct.C"
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//                 },
//                 "FileName": "files/zrealm_crossrealm21f.gno",
//                 "IsMethod": false,
//                 "Name": "cb",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/crossrealm_test",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "files/zrealm_crossrealm21f.gno",
//                         "Line": "8",
//                         "PkgPath": "gno.land/r/crossrealm_test"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "c",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.PointerType",
//                                 "Elt": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm/struct.C"
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
//                 "FileName": "files/zrealm_crossrealm21f.gno",
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
//                         "File": "files/zrealm_crossrealm21f.gno",
//                         "Line": "11",
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
// u[5965ae851c02ab677bc8394b408535c1db9b2635:9]={
//     "ObjectInfo": {
//         "ID": "5965ae851c02ab677bc8394b408535c1db9b2635:9",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "OwnerID": "5965ae851c02ab677bc8394b408535c1db9b2635:2",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests/crossrealm/struct.C"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "ObjectID": "5965ae851c02ab677bc8394b408535c1db9b2635:10"
//         }
//     }
// }
// switchrealm["gno.land/r/demo/tests/crossrealm/struct"]
// c[5965ae851c02ab677bc8394b408535c1db9b2635:14]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "d"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "5965ae851c02ab677bc8394b408535c1db9b2635:14",
//         "ModTime": "0",
//         "OwnerID": "5965ae851c02ab677bc8394b408535c1db9b2635:10",
//         "RefCount": "1"
//     }
// }
// u[5965ae851c02ab677bc8394b408535c1db9b2635:10]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/struct.A"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "77afd6587e56836f288520b275bebc4911b7ef57",
//                 "ObjectID": "5965ae851c02ab677bc8394b408535c1db9b2635:11"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/struct.B"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "c4e78b61af7cd5557452e3d22d20d1c3b1de4b41",
//                 "ObjectID": "5965ae851c02ab677bc8394b408535c1db9b2635:12"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/struct.D"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "70a58641e15bf62bf3d1d95ce2e8b2048286f42f",
//                 "ObjectID": "5965ae851c02ab677bc8394b408535c1db9b2635:14"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "5965ae851c02ab677bc8394b408535c1db9b2635:10",
//         "ModTime": "13",
//         "OwnerID": "5965ae851c02ab677bc8394b408535c1db9b2635:9",
//         "RefCount": "1"
//     }
// }
// d[5965ae851c02ab677bc8394b408535c1db9b2635:13]
// switchrealm["gno.land/r/crossrealm_test"]
