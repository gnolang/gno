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
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:12]={
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:12",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:11",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.PointerType",
//             "Elt": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm_b.fooer"
//             }
//         },
//         "V": {
//             "@type": "/gno.PointerValue",
//             "Base": {
//                 "@type": "/gno.RefValue",
//                 "Escaped": true,
//                 "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:4"
//             },
//             "Index": "0",
//             "TV": null
//         }
//     }
// }
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:11]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:11",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:6",
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
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "d8688b746889fec7c907a1a24e1958d728626994",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:12"
//             }
//         }
//     ]
// }
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:6]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:6",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//              "RefCount": "1"
//          },
//     @@ -9,6 +9,59 @@
//              "T": {
//                  "@type": "/gno.RefType",
//                  "ID": "gno.land/r/demo/tests/crossrealm.FooerGetter"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.FuncValue",
//     +            "Captures": [
//     +                {
//     +                    "T": {
//     +                        "@type": "/gno.heapItemType"
//     +                    },
//     +                    "V": {
//     +                        "@type": "/gno.RefValue",
//     +                        "Hash": "d8688b746889fec7c907a1a24e1958d728626994",
//     +                        "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:12"
//     +                    }
//     +                }
//     +            ],
//     +            "Closure": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "5d17d588a7dbca403fb3b6b9256705df64ce30e3",
//     +                "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:11"
//     +            },
//     +            "FileName": "",
//     +            "IsMethod": false,
//     +            "Name": "",
//     +            "NativeName": "",
//     +            "NativePkg": "",
//     +            "PkgPath": "gno.land/r/crossrealm_test",
//     +            "Source": {
//     +                "@type": "/gno.RefNode",
//     +                "BlockNode": null,
//     +                "Location": {
//     +                    "Column": "40",
//     +                    "File": "files/zrealm_crossrealm22.gno",
//     +                    "Line": "13",
//     +                    "PkgPath": "gno.land/r/crossrealm_test"
//     +                }
//     +            },
//     +            "SwitchRealm": false,
//     +            "Type": {
//     +                "@type": "/gno.FuncType",
//     +                "IsClosure": false,
//     +                "Params": null,
//     +                "Results": [
//     +                    {
//     +                        "Embedded": false,
//     +                        "Name": ".res.0",
//     +                        "Tag": "",
//     +                        "Type": {
//     +                            "@type": "/gno.RefType",
//     +                            "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//     +                        }
//     +                    }
//     +                ]
//     +            }
//              }
//          }
//      }
// u[0edc46caf30c00efd87b6c272673239eafbd051e:4]=
//     @@ -1,9 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:4",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "12",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
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
//     +        "ModTime": "12",
//     +        "RefCount": "3"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:5]=
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
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:5",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:4",
//              "RefCount": "1"
//          }
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:5]=
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
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:6]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:6",
//     -        "ModTime": "10",
//     +        "ModTime": "12",
//              "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//              "RefCount": "1"
//          },
//     @@ -12,37 +12,25 @@
//              },
//              "V": {
//                  "@type": "/gno.FuncValue",
//     -            "Captures": [
//     -                {
//     -                    "T": {
//     -                        "@type": "/gno.heapItemType"
//     -                    },
//     -                    "V": {
//     -                        "@type": "/gno.RefValue",
//     -                        "Hash": "d8688b746889fec7c907a1a24e1958d728626994",
//     -                        "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:12"
//     -                    }
//     -                }
//     -            ],
//                  "Closure": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5d17d588a7dbca403fb3b6b9256705df64ce30e3",
//     -                "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:11"
//     +                "Escaped": true,
//     +                "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:7"
//                  },
//                  "FileName": "",
//                  "IsMethod": false,
//                  "Name": "",
//                  "NativeName": "",
//                  "NativePkg": "",
//     -            "PkgPath": "gno.land/r/crossrealm_test",
//     +            "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
//                  "Source": {
//                      "@type": "/gno.RefNode",
//                      "BlockNode": null,
//                      "Location": {
//     -                    "Column": "40",
//     -                    "File": "files/zrealm_crossrealm22.gno",
//     -                    "Line": "13",
//     -                    "PkgPath": "gno.land/r/crossrealm_test"
//     +                    "Column": "23",
//     +                    "File": "crossrealm.gno",
//     +                    "Line": "25",
//     +                    "PkgPath": "gno.land/r/demo/tests/crossrealm_b"
//                      }
//                  },
//                  "SwitchRealm": false,
// u[0edc46caf30c00efd87b6c272673239eafbd051e:7]=
//     @@ -3,8 +3,8 @@
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:7",
//              "IsEscaped": true,
//     -        "ModTime": "0",
//     -        "RefCount": "5"
//     +        "ModTime": "12",
//     +        "RefCount": "6"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// u[0edc46caf30c00efd87b6c272673239eafbd051e:4]=
//     @@ -4,7 +4,7 @@
//              "IsEscaped": true,
//              "ModTime": "12",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "7c146728a91cd7c6aa3141a00cede434f9601daa",
//     +            "Hash": "c0786e5141354690b4ddd2792c20c7908301fba6",
//                  "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:5"
//              }
//          }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:3]=
//     @@ -4,7 +4,7 @@
//              "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//              "IsEscaped": true,
//              "ModTime": "12",
//     -        "RefCount": "3"
//     +        "RefCount": "2"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:11]
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:12]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:5]=
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
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:13]={
//     "Blank": {},
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:13",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:6",
//         "RefCount": "1"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "Escaped": true,
//         "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:7"
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
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:6]=
//     @@ -14,8 +14,8 @@
//                  "@type": "/gno.FuncValue",
//                  "Closure": {
//                      "@type": "/gno.RefValue",
//     -                "Escaped": true,
//     -                "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:7"
//     +                "Hash": "f46745f132c60b9ff79f65b23c29c68e8e5aa457",
//     +                "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:13"
//                  },
//                  "FileName": "",
//                  "IsMethod": false,
//     @@ -27,9 +27,9 @@
//                      "@type": "/gno.RefNode",
//                      "BlockNode": null,
//                      "Location": {
//     -                    "Column": "23",
//     +                    "Column": "62",
//                          "File": "crossrealm.gno",
//     -                    "Line": "25",
//     +                    "Line": "26",
//                          "PkgPath": "gno.land/r/demo/tests/crossrealm_b"
//                      }
//                  },
// u[0edc46caf30c00efd87b6c272673239eafbd051e:7]=(noop)
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
