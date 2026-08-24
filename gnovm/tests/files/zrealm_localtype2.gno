// PKGPATH: gno.land/r/test
// Mirrored (plus restart coverage) by restart_local_type.txtar's zltsh realm.
package test

type T struct{ x int }

func (t T) Get() int { return t.x }

type Getter interface{ Get() int }

var (
	G      func() any
	GG     func() any
	D1, D2 any
)

func mk1() any {
	type S struct{ T }
	return S{T{1}}
}

func mk2() any {
	type S struct{ T }
	return S{T{2}}
}

func main(cur realm) {
	// B: no-trace escape — nothing S2-typed persists at save time, not even
	// a TypeValue (the type name is preprocess-folded, not captured; the
	// golden shows only the FuncValue). Reload re-mints S2, same TypeID.
	type S2 struct{ T }
	G = func() any { return S2{T{7}} }
	println(G().(Getter).Get())

	// C: type declared inside a closure literal (ParentLoc = FuncLitExpr)
	GG = func() any {
		type C struct{ T }
		return C{T{5}}
	}
	println(GG().(Getter).Get())

	// D: same local type name in two different functions
	D1, D2 = mk1(), mk2()
	println(D1.(Getter).Get(), D2.(Getter).Get())
}

// Output:
// 7
// 5
// 1 2

// Realm:
// finalizerealm["gno.land/r/test"]
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:11](319)={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": true,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "LastObjectSize": "319",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/test",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "zrealm_localtype2.gno",
//             "PkgPath": "gno.land/r/test",
//             "Span": {
//                 "End": {
//                     "Column": "36",
//                     "Line": "32"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "6",
//                     "Line": "32"
//                 }
//             }
//         }
//     },
//     "Type": {
//         "@type": "/gno.FuncType",
//         "Params": [],
//         "Results": [
//             {
//                 "Embedded": false,
//                 "Name": ".res.0",
//                 "PkgPath": "",
//                 "Tag": "",
//                 "Type": {
//                     "@type": "/gno.InterfaceType",
//                     "Generic": "",
//                     "Methods": [],
//                     "PkgPath": ""
//                 }
//             }
//         ]
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:12](319)={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": true,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "LastObjectSize": "319",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/test",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "zrealm_localtype2.gno",
//             "PkgPath": "gno.land/r/test",
//             "Span": {
//                 "End": {
//                     "Column": "3",
//                     "Line": "39"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "7",
//                     "Line": "36"
//                 }
//             }
//         }
//     },
//     "Type": {
//         "@type": "/gno.FuncType",
//         "Params": [],
//         "Results": [
//             {
//                 "Embedded": false,
//                 "Name": ".res.0",
//                 "PkgPath": "",
//                 "Tag": "",
//                 "Type": {
//                     "@type": "/gno.InterfaceType",
//                     "Generic": "",
//                     "Methods": [],
//                     "PkgPath": ""
//                 }
//             }
//         ]
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:14](216)={
//     "Fields": [
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:13](333)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test.T"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "7ea4994d2e81627502d7d6e1f981f73c7ee052f6",
//                 "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13",
//         "LastObjectSize": "333",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:16](216)={
//     "Fields": [
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:16",
//         "LastObjectSize": "216",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:15",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:15](333)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test.T"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "87820fcd0f156dec9d5674e0fafdfa14cf6ad99d",
//                 "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:16"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:15",
//         "LastObjectSize": "333",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "RefCount": "1"
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](112)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "241",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -24,6 +24,11 @@
//                          }
//                      }
//                  ]
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "ca70cd14ed4a35a71307427fc2989fa0f6c5f2f6",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](112)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "241",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -24,6 +24,11 @@
//                          }
//                      }
//                  ]
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "78a9408acaf9be2846a14ef73b18d678db182cc6",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:5](198)=
//     @@ -2,9 +2,19 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.RefType",
//     +            "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype2.gno:17:1-20:2].S"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "6391232219e0e9a1dfde585eb3a0ad0cec569dc6",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13"
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:6](198)=
//     @@ -2,9 +2,19 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//              "LastObjectSize": "190",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.RefType",
//     +            "ID": "gno.land/r/test[gno.land/r/test/zrealm_localtype2.gno:22:1-25:2].S"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "a487e2e9231b9867a114d158c1e35a7467dc61b6",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:15"
//     +        }
//     +    }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "1470",
//     -        "ModTime": "0",
//     +        "ModTime": "16",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -58,7 +58,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e457ee3297d44037f86e277b030ba75d1cc77177",
//     +                "Hash": "38b8b0228ed46e13b122cb043e386c585ed78f50",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
//     @@ -68,7 +68,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "ed5c72731df782c2a7a498f131024a7e8abc1277",
//     +                "Hash": "1c645eb480e10c5ab201117bd7e9e105a5704339",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                  }
//              },
//     @@ -78,7 +78,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "091813d68ddddd3d9a460952d01c6e8427f33f07",
//     +                "Hash": "d0b0e7c20d8a96c64d14bac1b10a19d4d7e55723",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:5"
//                  }
//              },
//     @@ -88,7 +88,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "50a7f52f8f9d6e7936afc5575fedbf44ca9b5794",
//     +                "Hash": "715e82ae508b7982380555216ccd8f9d743f3919",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6"
//                  }
//              },
