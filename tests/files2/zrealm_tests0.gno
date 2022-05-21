// PKGPATH: gno.land/r/tests_test
package tests_test

import (
	"gno.land/r/tests"
	"gno.land/r/tests_foo"
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
// switchrealm["gno.land/r/tests"]
// c[adc8ca1e1018f3da5f2caaf455c39e10cc1284db:10]={
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
//         "ID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:10",
//         "ModTime": "0",
//         "OwnerID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:9",
//         "RefCount": "1"
//     }
// }
// c[adc8ca1e1018f3da5f2caaf455c39e10cc1284db:9]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/tests_foo.FooStringer"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "40af94627050bc75188e7e3bb27ad1f531552093",
//                         "ObjectID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:6"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/tests_foo.FooStringer"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "ad44c3b08de7a27be4e1582353e4dbd0e371ef37",
//                         "ObjectID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:8"
//                     }
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": null,
//                 "Index": "0",
//                 "TV": {
//                     "T": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/tests_foo.FooStringer"
//                     },
//                     "V": {
//                         "@type": "/gno.RefValue",
//                         "Hash": "7c642db97643cd0efa50d6a3387d05273c9e90bf",
//                         "ObjectID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:10"
//                     }
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:9",
//         "ModTime": "0",
//         "OwnerID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:2",
//         "RefCount": "1"
//     }
// }
// u[adc8ca1e1018f3da5f2caaf455c39e10cc1284db:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:2",
//         "IsEscaped": true,
//         "ModTime": "8",
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
//             "PkgPath": "gno.land/r/tests"
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
//                         "PkgPath": "gno.land/r/tests"
//                     },
//                     "Methods": [],
//                     "Name": "Stringer",
//                     "PkgPath": "gno.land/r/tests"
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
//                         "PkgPath": "gno.land/r/tests"
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
//                                                 "ID": "gno.land/r/tests.TestRealmObject"
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
//                                 "PkgPath": "gno.land/r/tests",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "tests.gno",
//                                         "Line": "20",
//                                         "Nonce": "0",
//                                         "PkgPath": "gno.land/r/tests"
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
//                                                     "ID": "gno.land/r/tests.TestRealmObject"
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
//                                         "Name": "t",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.PointerType",
//                                             "Elt": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/tests.TestRealmObject"
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
//                                 "PkgPath": "gno.land/r/tests",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "File": "tests.gno",
//                                         "Line": "20",
//                                         "Nonce": "0",
//                                         "PkgPath": "gno.land/r/tests"
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
//                                                     "ID": "gno.land/r/tests.TestRealmObject"
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
//                     "PkgPath": "gno.land/r/tests"
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
//                                         "ID": "gno.land/r/tests.TestNode"
//                                     }
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/tests"
//                     },
//                     "Methods": [],
//                     "Name": "TestNode",
//                     "PkgPath": "gno.land/r/tests"
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
//                             "ID": "gno.land/r/tests.Stringer"
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
//                     "ObjectID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:3"
//                 },
//                 "FileName": "interfaces.gno",
//                 "IsMethod": false,
//                 "Name": "AddStringer",
//                 "PkgPath": "gno.land/r/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "interfaces.gno",
//                         "Line": "13",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/tests"
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
//                                 "ID": "gno.land/r/tests.Stringer"
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
//                     "ObjectID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:3"
//                 },
//                 "FileName": "interfaces.gno",
//                 "IsMethod": false,
//                 "Name": "Render",
//                 "PkgPath": "gno.land/r/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "interfaces.gno",
//                         "Line": "20",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/tests"
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
//                     "ObjectID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "CurrentRealmPath",
//                 "PkgPath": "gno.land/r/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "5",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/tests"
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
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "t",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.PointerType",
//                             "Elt": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/tests.TestRealmObject"
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
//                     "ObjectID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "ModifyTestRealmObject",
//                 "PkgPath": "gno.land/r/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "16",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/tests"
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
//                                     "ID": "gno.land/r/tests.TestRealmObject"
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
//                     "ObjectID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "InitTestNodes",
//                 "PkgPath": "gno.land/r/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "36",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/tests"
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
//                     "ObjectID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "ModTestNodes",
//                 "PkgPath": "gno.land/r/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "41",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/tests"
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
//                     "ObjectID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:4"
//                 },
//                 "FileName": "tests.gno",
//                 "IsMethod": false,
//                 "Name": "PrintTestNodes",
//                 "PkgPath": "gno.land/r/tests",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "File": "tests.gno",
//                         "Line": "49",
//                         "Nonce": "0",
//                         "PkgPath": "gno.land/r/tests"
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
//                 "@type": "/gno.SliceType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests.Stringer"
//                 },
//                 "Vrd": false
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "2338c6d54a0e93b23daebd4988eafe493e1824fa",
//                     "ObjectID": "adc8ca1e1018f3da5f2caaf455c39e10cc1284db:9"
//                 },
//                 "Length": "3",
//                 "Maxcap": "3",
//                 "Offset": "0"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests.TestNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests.TestNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests.TestNode"
//                 }
//             }
//         }
//     ]
// }
// d[adc8ca1e1018f3da5f2caaf455c39e10cc1284db:7]
// switchrealm["gno.land/r/tests_foo"]
// switchrealm["gno.land/r/tests_foo"]
// switchrealm["gno.land/r/tests_foo"]
// switchrealm["gno.land/r/tests_foo"]
// switchrealm["gno.land/r/tests"]
// switchrealm["gno.land/r/tests_test"]
