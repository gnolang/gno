// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
)

func main() {
	switchrealm()

	f := crossrealm_b.Fooer
	withswitch(crossrealm.SetFooerGetter)(func() crossrealm.Fooer { return f })
	withswitch(crossrealm.CallFooerGetterFooSR2)()
	f.SetS("B")
	withswitch(crossrealm.CallFooerGetterFooSR2)()
	println(".")

	f.SetS("C")
	withswitch(crossrealm.SetFooerGetter)(crossrealm_b.FooerGetter)
	withswitch(crossrealm.CallFooerGetterFooSR2)()
	println(".")

	f.SetS("D")
	withswitch(crossrealm.SetFooerGetter)(crossrealm_b.FooerGetterBuilder())
	withswitch(crossrealm.CallFooerGetterFooSR2)()
	println(".")
}

// Output:
// hello A cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// hello B cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .
// hello C cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .
// hello D cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:7]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:7",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//         "RefCount": "1"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "1",
//             "File": "files/zrealm_crossrealm22.gno",
//             "Line": "9",
//             "PkgPath": "gno.land/r/crossrealm_test"
//         }
//     },
//     "Values": [
//         {
//             "N": "UmVhRG9OTFk=",
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm_b.fooer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:3"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ]
// }
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:2]=
//     @@ -3,7 +3,7 @@
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//              "IsEscaped": true,
//     -        "ModTime": "4",
//     +        "ModTime": "6",
//              "RefCount": "3"
//          },
//          "Parent": null,
//     @@ -574,6 +574,47 @@
//                  "T": {
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/demo/tests/crossrealm.FooerGetter"
//     +            },
//     +            "V": {
//     +                "@type": "/gno.FuncValue",
//     +                "Closure": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "bb158c8b1419261454c379d1e4da78f87f602847",
//     +                    "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:7"
//     +                },
//     +                "FileName": "",
//     +                "IsMethod": false,
//     +                "Name": "",
//     +                "NativeName": "",
//     +                "NativePkg": "",
//     +                "PkgPath": "gno.land/r/crossrealm_test",
//     +                "Source": {
//     +                    "@type": "/gno.RefNode",
//     +                    "BlockNode": null,
//     +                    "Location": {
//     +                        "Column": "40",
//     +                        "File": "files/zrealm_crossrealm22.gno",
//     +                        "Line": "13",
//     +                        "PkgPath": "gno.land/r/crossrealm_test"
//     +                    }
//     +                },
//     +                "SwitchRealm": false,
//     +                "Type": {
//     +                    "@type": "/gno.FuncType",
//     +                    "IsClosure": false,
//     +                    "Params": null,
//     +                    "Results": [
//     +                        {
//     +                            "Embedded": false,
//     +                            "Name": "",
//     +                            "Tag": "",
//     +                            "Type": {
//     +                                "@type": "/gno.RefType",
//     +                                "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//     +                            }
//     +                        }
//     +                    ]
//     +                }
//                  }
//              },
//              {
// u[0edc46caf30c00efd87b6c272673239eafbd051e:3]=
//     @@ -1,9 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "7",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:2",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[f5a516808f8976c33939133293d598ce3bca4e8d:3]=
//     @@ -3,8 +3,8 @@
//          "ObjectInfo": {
//              "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//              "IsEscaped": true,
//     -        "ModTime": "0",
//     -        "RefCount": "2"
//     +        "ModTime": "7",
//     +        "RefCount": "3"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:4]=
//     @@ -7,13 +7,13 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "A"
//     +                "value": "B"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:4",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
//              "RefCount": "1"
//          }
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:4]=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "B"
//     +                "value": "C"
//                  }
//              }
//          ],
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:2]=
//     @@ -3,7 +3,7 @@
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//              "IsEscaped": true,
//     -        "ModTime": "6",
//     +        "ModTime": "7",
//              "RefCount": "3"
//          },
//          "Parent": null,
//     @@ -579,23 +579,23 @@
//                      "@type": "/gno.FuncValue",
//                      "Closure": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "bb158c8b1419261454c379d1e4da78f87f602847",
//     -                    "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:7"
//     +                    "Escaped": true,
//     +                    "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:5"
//                      },
//                      "FileName": "",
//                      "IsMethod": false,
//                      "Name": "",
//                      "NativeName": "",
//                      "NativePkg": "",
//     -                "PkgPath": "gno.land/r/crossrealm_test",
//     -                "Source": {
//     -                    "@type": "/gno.RefNode",
//     -                    "BlockNode": null,
//     -                    "Location": {
//     -                        "Column": "40",
//     -                        "File": "files/zrealm_crossrealm22.gno",
//     -                        "Line": "13",
//     -                        "PkgPath": "gno.land/r/crossrealm_test"
//     +                "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
//     +                "Source": {
//     +                    "@type": "/gno.RefNode",
//     +                    "BlockNode": null,
//     +                    "Location": {
//     +                        "Column": "23",
//     +                        "File": "crossrealm.gno",
//     +                        "Line": "25",
//     +                        "PkgPath": "gno.land/r/demo/tests/crossrealm_b"
//                          }
//                      },
//                      "SwitchRealm": false,
// u[0edc46caf30c00efd87b6c272673239eafbd051e:5]=
//     @@ -3,8 +3,8 @@
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:5",
//              "IsEscaped": true,
//     -        "ModTime": "0",
//     -        "RefCount": "3"
//     +        "ModTime": "7",
//     +        "RefCount": "4"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// u[0edc46caf30c00efd87b6c272673239eafbd051e:3]=
//     @@ -4,7 +4,7 @@
//              "IsEscaped": true,
//              "ModTime": "7",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:2",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "8222b763e9bc04b4b7805e165e9f1324a39f28b6",
//     +            "Hash": "759fd5b507fff8ea1b18d401550d918387a63445",
//                  "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:4"
//              }
//          }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:3]=
//     @@ -4,7 +4,7 @@
//              "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//              "IsEscaped": true,
//              "ModTime": "7",
//     -        "RefCount": "3"
//     +        "RefCount": "2"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:7]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:4]=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "C"
//     +                "value": "D"
//                  }
//              }
//          ],
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:8]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:8",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//         "RefCount": "1"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:5"
//     },
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "23",
//             "File": "crossrealm.gno",
//             "Line": "26",
//             "PkgPath": "gno.land/r/demo/tests/crossrealm_b"
//         }
//     },
//     "Values": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm.FooerGetter"
//             }
//         }
//     ]
// }
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:2]=
//     @@ -579,8 +579,8 @@
//                      "@type": "/gno.FuncValue",
//                      "Closure": {
//                          "@type": "/gno.RefValue",
//     -                    "Escaped": true,
//     -                    "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:5"
//     +                    "Hash": "f6d4d9c32bce586686df77b6c915950f10c94ede",
//     +                    "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:8"
//                      },
//                      "FileName": "",
//                      "IsMethod": false,
//     @@ -592,9 +592,9 @@
//                          "@type": "/gno.RefNode",
//                          "BlockNode": null,
//                          "Location": {
//     -                        "Column": "23",
//     -                        "File": "crossrealm.gno",
//     -                        "Line": "25",
//     +                        "Column": "62",
//     +                        "File": "crossrealm.gno",
//     +                        "Line": "26",
//                              "PkgPath": "gno.land/r/demo/tests/crossrealm_b"
//                          }
//                      },
// u[0edc46caf30c00efd87b6c272673239eafbd051e:5]=(noop)
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
