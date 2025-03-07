// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/iface"
)

type fooer struct {
	name string
	bar
}
type bar struct{ name string }

func (f fooer) Foo() { println("hello " + f.bar.name) }

var f = fooer{name: "local_fooer", bar: bar{name: "local_bar_1"}}
var b = bar{name: "local_bar_2"} // real

func main() {
	f.bar = b // this makes f dirty
	crossrealm.SetContainer2(&f)
	crossrealm.RunF()
	print(".")
}

// Output:
// hello local_bar_2
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/iface"]
// u[5836d1a3cec217737a5553d4c1a03d8897d5f281:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
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
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:2"
//                 },
//                 "Index": "2",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "5836d1a3cec217737a5553d4c1a03d8897d5f281:4",
//         "ModTime": "4",
//         "OwnerID": "5836d1a3cec217737a5553d4c1a03d8897d5f281:2",
//         "RefCount": "1"
//     }
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:2]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
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
//             "PkgPath": "gno.land/r/crossrealm_test"
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
//                             },
//                             {
//                                 "Embedded": true,
//                                 "Name": "bar",
//                                 "Tag": "",
//                                 "Type": {
//                                     "@type": "/gno.RefType",
//                                     "ID": "gno.land/r/crossrealm_test.bar"
//                                 }
//                             }
//                         ],
//                         "PkgPath": "gno.land/r/crossrealm_test"
//                     },
//                     "Methods": [
//                         {
//                             "T": {
//                                 "@type": "/gno.FuncType",
//                                 "IsClosure": false,
//                                 "Params": [
//                                     {
//                                         "Embedded": false,
//                                         "Name": "f",
//                                         "Tag": "",
//                                         "Type": {
//                                             "@type": "/gno.RefType",
//                                             "ID": "gno.land/r/crossrealm_test.fooer"
//                                         }
//                                     }
//                                 ],
//                                 "Results": []
//                             },
//                             "V": {
//                                 "@type": "/gno.FuncValue",
//                                 "Closure": null,
//                                 "FileName": "files/zrealm_crossrealm21e.gno",
//                                 "IsMethod": true,
//                                 "Name": "Foo",
//                                 "NativeName": "",
//                                 "NativePkg": "",
//                                 "PkgPath": "gno.land/r/crossrealm_test",
//                                 "Source": {
//                                     "@type": "/gno.RefNode",
//                                     "BlockNode": null,
//                                     "Location": {
//                                         "Column": "1",
//                                         "File": "files/zrealm_crossrealm21e.gno",
//                                         "Line": "14",
//                                         "PkgPath": "gno.land/r/crossrealm_test"
//                                     }
//                                 },
//                                 "Type": {
//                                     "@type": "/gno.FuncType",
//                                     "IsClosure": false,
//                                     "Params": [
//                                         {
//                                             "Embedded": false,
//                                             "Name": "f",
//                                             "Tag": "",
//                                             "Type": {
//                                                 "@type": "/gno.RefType",
//                                                 "ID": "gno.land/r/crossrealm_test.fooer"
//                                             }
//                                         }
//                                     ],
//                                     "Results": []
//                                 }
//                             }
//                         }
//                     ],
//                     "Name": "fooer",
//                     "PkgPath": "gno.land/r/crossrealm_test"
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
//                         "PkgPath": "gno.land/r/crossrealm_test"
//                     },
//                     "Methods": [],
//                     "Name": "bar",
//                     "PkgPath": "gno.land/r/crossrealm_test"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/crossrealm_test.fooer"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/crossrealm_test.bar"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "f7273b4e66b51d48a2dbc66bc02e80c0262ccdd1",
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:5"
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
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:6"
//                 },
//                 "FileName": "files/zrealm_crossrealm21e.gno",
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
//                         "File": "files/zrealm_crossrealm21e.gno",
//                         "Line": "19",
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
//                 "value": "local_bar_2"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:7",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "RefCount": "1"
//     }
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:3]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "local_fooer"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/crossrealm_test.bar"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "2653fb8c9edc489f07585ec600dfd1a968d81ca1",
//                 "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:7"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "ModTime": "6",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "1"
//     }
// }
// d[f5a516808f8976c33939133293d598ce3bca4e8d:4]
// switchrealm["gno.land/r/demo/tests/crossrealm/iface"]
// switchrealm["gno.land/r/crossrealm_test"]
