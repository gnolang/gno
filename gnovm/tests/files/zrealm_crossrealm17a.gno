// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm"
)

type container struct {
	name string
	f    *fooer
}

func (container) Foo() { println("hello " + c.f.name) }

type fooer struct{ name string }

var c = &container{name: "local_container"}

func main() {
	ff := &fooer{name: "local_foo"}
	c.f = ff
	crossrealm.SetFooer(c)
	crossrealm.CallFooerFoo()
	print(".")
}

// Output:
// hello local_foo
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//         "IsEscaped": true,
//         "ModTime": "5",
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
//                                 "IsClosure": false,
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
//                                     "IsClosure": false,
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
//                     "Hash": "a75fdb389fedfcbbaa7f446d528c1e149726347c",
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:4"
//                 },
//                 "Index": "0",
//                 "TV": null
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:3"
//                 },
//                 "FileName": "crossrealm_p.gno",
//                 "IsMethod": false,
//                 "Name": "init.2",
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:3"
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
//                     "IsClosure": false,
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
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.container"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:3"
//                 },
//                 "FileName": "crossrealm_p.gno",
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
//                         "File": "crossrealm_p.gno",
//                         "Line": "35",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                 "IsClosure": false,
//                 "Params": [],
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
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:3"
//                 },
//                 "FileName": "crossrealm_p.gno",
//                 "IsMethod": false,
//                 "Name": "GetFooer",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_p.gno",
//                         "Line": "40",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                 "IsClosure": false,
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
//                 "FileName": "crossrealm_p.gno",
//                 "IsMethod": false,
//                 "Name": "CallFooerFoo",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_p.gno",
//                         "Line": "42",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                                 }
//                             }
//                         ]
//                     },
//                     "Methods": [],
//                     "Name": "FooerGetter",
//                     "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.FooerGetter"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.FuncType",
//                 "IsClosure": false,
//                 "Params": [
//                     {
//                         "Embedded": false,
//                         "Name": "fg",
//                         "Tag": "",
//                         "Type": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests/crossrealm.FooerGetter"
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
//                             "ID": "gno.land/r/demo/tests/crossrealm.FooerGetter"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:3"
//                 },
//                 "FileName": "crossrealm_p.gno",
//                 "IsMethod": false,
//                 "Name": "SetFooerGetter",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_p.gno",
//                         "Line": "48",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
//                     }
//                 },
//                 "Type": {
//                     "@type": "/gno.FuncType",
//                     "IsClosure": false,
//                     "Params": [
//                         {
//                             "Embedded": false,
//                             "Name": "fg",
//                             "Tag": "",
//                             "Type": {
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.FooerGetter"
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
//                                 "ID": "gno.land/r/demo/tests/crossrealm.FooerGetter"
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
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/demo/tests/crossrealm.FooerGetter"
//                         }
//                     }
//                 ]
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:3"
//                 },
//                 "FileName": "crossrealm_p.gno",
//                 "IsMethod": false,
//                 "Name": "GetFooerGetter",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_p.gno",
//                         "Line": "53",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//                                 "@type": "/gno.RefType",
//                                 "ID": "gno.land/r/demo/tests/crossrealm.FooerGetter"
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
//                 "Results": []
//             },
//             "V": {
//                 "@type": "/gno.FuncValue",
//                 "Closure": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:3"
//                 },
//                 "FileName": "crossrealm_p.gno",
//                 "IsMethod": false,
//                 "Name": "CallFooerGetterFoo",
//                 "NativeName": "",
//                 "NativePkg": "",
//                 "PkgPath": "gno.land/r/demo/tests/crossrealm",
//                 "Source": {
//                     "@type": "/gno.RefNode",
//                     "BlockNode": null,
//                     "Location": {
//                         "Column": "1",
//                         "File": "crossrealm_p.gno",
//                         "Line": "57",
//                         "PkgPath": "gno.land/r/demo/tests/crossrealm"
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
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "IsEscaped": true,
//         "ModTime": "5",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.container"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//         }
//     }
// }
// switchrealm["gno.land/r/crossrealm_test"]
// c[f5a516808f8976c33939133293d598ce3bca4e8d:7]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "local_foo"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:7",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:6",
//         "RefCount": "1"
//     }
// }
// c[f5a516808f8976c33939133293d598ce3bca4e8d:6]={
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:6",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.fooer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "90e6414435c56b7ffa17bd575535b7a558b93e22",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:7"
//         }
//     }
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "local_container"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.fooer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "7d3bda685439b3989f18e824762f473f40c944fb",
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:6"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "ModTime": "5",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
