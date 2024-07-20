// PKGPATH: gno.land/r/demo/tests_test
package tests_test

import (
	"gno.land/r/demo/tests"
	"gno.land/r/demo/tests_foo"
)

func init() {
	tests_foo.AddFooStringer("one")
	tests_foo.AddFooStringer("two")
}

func main() {
	tests_foo.AddFooStringer("three")
	println(tests.Render(""))
}

// Output:
// 0: &FooStringer{one}
// 1: &FooStringer{two}
// 2: &FooStringer{three}

// Realm:
// switchrealm["gno.land/r/demo/tests"]
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "three"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17",
//         "RefCount": "1"
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17]={
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:16",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "d3d6ffa52602f2bc976051d79294d219750aca64",
//             "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:18"
//         }
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:16]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "4ea1e08156f3849b74a0f41f92cd4b48fb94926b",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11"
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
//                     "ID": "gno.land/r/demo/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "ce86ea1156e75a44cd9d7ba2261819b100aa4ed1",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:14"
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
//                     "ID": "gno.land/r/demo/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "b66192fbd8a8dde79b6f854b5cc3c4cc965cfd92",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:17"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:16",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2",
//         "RefCount": "1"
//     }
// }
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2",
//         "IsEscaped": true,
//         "ModTime": "15",
//         "RefCount": "5"
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
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests.Stringer"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "ad25f70f66c8c53042afd1377e5ff5ab744bf1a5",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:16"
//                 },
//                 "Length": "3",
//                 "Maxcap": "3",
//                 "Offset": "0"
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
//                 "NativeName": "",
//                 "NativePkg": "",
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
//                 "NativeName": "",
//                 "NativePkg": "",
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
//                 "@type": "/gno.TypeType"
//             },
//             "V": {
//                 "@type": "/gno.TypeValue",
//                 "Type": {
//                     "@type": "/gno.DeclaredType",
//                     "Base": {
//                         "@type": "/gno.PrimitiveType",
//                         "value": "2048"
//                     },
//                     "Methods": [],
//                     "Name": "Word",
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
//                         "@type": "/gno.SliceType",
//                         "Elt": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests.Word"
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
//                                         "Name": "n",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "gno.land/r/demo/tests.nat"
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
//                                             "ID": "gno.land/r/demo/tests.nat"
//                                         }
//                                     }
//                                 ]
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "realm_method38d.gno",
//                                 "IsMethod": true,
//                                 "Name": "Add",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/demo/tests",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "realm_method38d.gno",
//                                         "Line": "5",
//                                         "Nonce": "0",
//                                         "PkgPath": "gno.land/r/demo/tests"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "n",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/demo/tests.nat"
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
//                                                 "ID": "gno.land/r/demo/tests.nat"
//                                             }
//                                         }
//                                     ]
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "nat",
//                     "PkgPath": "gno.land/r/demo/tests"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests.Int"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "49283398258e135138cd8e234142d5daaa8c661d",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "Index": "0",
//                 "TV": null
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
//                                 "Name": "neg",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.PrimitiveType",
//                                     "value": "4"
//                                 }
//                             },
//                             {
//                                 "Embedded": false,
//                                 "Name": "abs",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests.nat"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/demo/tests"
//                     },
//                     "Methods": [],
//                     "Name": "Int",
//                     "PkgPath": "gno.land/r/demo/tests"
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
//                             "ID": "gno.land/r/demo/tests.nat"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:7"
//                 },
//                 "FileName": "realm_compositelit.gno",
//                 "IsMethod": false,
//                 "Name": "GetZeroType",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "realm_compositelit.gno",
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
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests.nat"
//                             }
//                         }
//                     ]
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests.nat"
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
//                             "ID": "gno.land/r/demo/tests.nat"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:8"
//                 },
//                 "FileName": "realm_method38d.gno",
//                 "IsMethod": false,
//                 "Name": "GetAbs",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "realm_method38d.gno",
//                         "Line": "9",
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
//                                 "ID": "gno.land/r/demo/tests.nat"
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
//                             "ID": "gno.land/r/demo/tests.nat"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:8"
//                 },
//                 "FileName": "realm_method38d.gno",
//                 "IsMethod": false,
//                 "Name": "AbsAdd",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "realm_method38d.gno",
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
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests.nat"
//                             }
//                         }
//                     ]
//                 }
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
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "IncCounter",
//                 "NativeName": "",
//                 "NativePkg": "",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "Counter",
//                 "NativeName": "",
//                 "NativePkg": "",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "CurrentRealmPath",
//                 "NativeName": "",
//                 "NativePkg": "",
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
//                 "@type": "/gno.RefType",
//                 "ID": "std.Address"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm"
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
//                             "ID": "std.Address"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "InitOrigCaller",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "25",
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
//                                 "ID": "std.Address"
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "CallAssertOriginCall",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "29",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "CallIsOriginCall",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "33",
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
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "CallSubtestsAssertOriginCall",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "37",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "CallSubtestsIsOriginCall",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "41",
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
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/demo/tests",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "tests.gno",
//                                         "Line": "56",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "ModifyTestRealmObject",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "52",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "InitTestNodes",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "74",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "ModTestNodes",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "79",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "PrintTestNodes",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "87",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "GetPrevRealm",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "91",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "GetRSubtestsPrevRealm",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "95",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "Exec",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "99",
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
//         }
//     ]
// }
// d[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:13]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests"]
// switchrealm["gno.land/r/demo/tests_test"]
