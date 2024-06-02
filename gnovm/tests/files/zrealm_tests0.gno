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
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:13]={
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
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:13",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12",
//         "RefCount": "1"
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12]={
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "3d98cb07a727c6d010396c5a28713b138fae7e62",
//             "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:13"
//         }
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11]={
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
//                     "Hash": "cc3a93e11a02d37a068cdb69187cd4c6e4dc294e",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:6"
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
//                     "Hash": "ce72679434389344e1775636b73148438153b314",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:9"
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
//                     "Hash": "115838ed3ba1e3d03fb7742aac6ed3635c9a8bf9",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11",
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
//         "ModTime": "10",
//         "RefCount": "3"
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
//                                         "Line": "42",
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
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
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
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "AssertOriginCall",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "23",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "IsOriginCall",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "27",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
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
//                         "Line": "38",
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
//                 "@type": "/gno.FuncType",
//                 "Params": [],
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
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
//                         "Line": "60",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
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
//                         "Line": "65",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
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
//                         "Line": "73",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
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
//                         "Line": "77",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
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
//                         "Line": "81",
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
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:4"
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
//                         "Line": "85",
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
//                     "Hash": "d26b76f9d5b71eeb6c02e3dcb3fe48de7179a6a6",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:11"
//                 },
//                 "Length": "3",
//                 "Maxcap": "3",
//                 "Offset": "0"
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
//         }
//     ]
// }
// d[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:8]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests"]
// switchrealm["gno.land/r/demo/tests_test"]
