// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
)

func main(cur realm) {
	f := crossrealm_b.Fooer
	crossrealm.SetFooerGetter(cross, func() crossrealm.Fooer { return f })
	crossrealm.CallFooerGetterFooCross(cross)
	f.SetS("B")
	crossrealm.CallFooerGetterFooCross(cross)
	println(".")

	f.SetS("C")
	crossrealm.SetFooerGetter(cross, crossrealm_b.FooerGetter)
	crossrealm.CallFooerGetterFooCross(cross)
	println(".")

	f.SetS("D")
	crossrealm.SetFooerGetter(cross, crossrealm_b.FooerGetterBuilder())
	crossrealm.CallFooerGetterFooCross(cross)
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
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:40]={
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:40",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:39",
//         "RefCount": "1"
//     },
//     "Value": {
//         "N": "UmVhRG9OTFk=",
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
//                 "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:7"
//             },
//             "Index": "0",
//             "TV": null
//         }
//     }
// }
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:39]={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "16ce0cb36c37f36d30ba2854375415c61fe647da",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:40"
//             }
//         }
//     ],
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:39",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:14",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/crossrealm_test",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "zrealm_crossrealm22.gno",
//             "PkgPath": "gno.land/r/crossrealm_test",
//             "Span": {
//                 "End": {
//                     "Column": "71",
//                     "Line": "11"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "35",
//                     "Line": "11"
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
//                 "Tag": "",
//                 "Type": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                 }
//             }
//         ]
//     }
// }
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:14]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:14",
//     -        "ModTime": "0",
//     +        "ModTime": "38",
//              "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//              "RefCount": "1"
//          },
//     @@ -9,6 +9,11 @@
//              "T": {
//                  "@type": "/gno.RefType",
//                  "ID": "gno.land/r/demo/tests/crossrealm.FooerGetter"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "80e056cdfbc5bfe285bef41c38aea25054d04ed8",
//     +            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:39"
//              }
//          }
//      }
// u[0edc46caf30c00efd87b6c272673239eafbd051e:7]=
//     @@ -1,9 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:7",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "40",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:6",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:8]=
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
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:8",
//     -        "ModTime": "0",
//     +        "ModTime": "17",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:7",
//              "RefCount": "1"
//          }
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:8]=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "B"
//     +                "value": "C"
//                  }
//              }
//          ],
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:14]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:14",
//     -        "ModTime": "38",
//     +        "ModTime": "40",
//              "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//              "RefCount": "1"
//          },
//     @@ -12,8 +12,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "80e056cdfbc5bfe285bef41c38aea25054d04ed8",
//     -            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:39"
//     +            "Escaped": true,
//     +            "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:10"
//              }
//          }
//      }
// u[0edc46caf30c00efd87b6c272673239eafbd051e:10]=
//     @@ -8,9 +8,10 @@
//          "NativePkg": "",
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:10",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "40",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:9",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
// u[0edc46caf30c00efd87b6c272673239eafbd051e:7]=
//     @@ -4,7 +4,7 @@
//              "IsEscaped": true,
//              "ModTime": "40",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:6",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "f72c517435a3de33aab61581c6b12aa6c37e7a7e",
//     +            "Hash": "5ead7f6dd6fe8f4da927938f8cb267570a233889",
//                  "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:8"
//              }
//          }
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:39]
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:40]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:8]=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "C"
//     +                "value": "D"
//                  }
//              }
//          ],
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:41]={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:41",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:14",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "crossrealm.gno",
//             "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
//             "Span": {
//                 "End": {
//                     "Column": "102",
//                     "Line": "36"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "62",
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
//                 "Tag": "",
//                 "Type": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests/crossrealm.Fooer"
//                 }
//             }
//         ]
//     }
// }
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:14]=
//     @@ -12,8 +12,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Escaped": true,
//     -            "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:10"
//     +            "Hash": "1d3abf2c5837873f853251b344bfa68131f8179d",
//     +            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:41"
//              }
//          }
//      }
// u[0edc46caf30c00efd87b6c272673239eafbd051e:10]=
//     @@ -11,7 +11,7 @@
//              "IsEscaped": true,
//              "ModTime": "40",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:9",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/crossrealm_test"]
