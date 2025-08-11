// PKGPATH: gno.land/r/crossrealm
package crossrealm

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
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:49](406)={
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:49",
//         "LastObjectSize": "406",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:48",
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
//                 "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:4"
//             },
//             "Index": "0",
//             "TV": null
//         }
//     }
// }
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:48](509)={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "20f512387a78830fb7ba6eead34322a272ae82f9",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:49"
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
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:48",
//         "LastObjectSize": "509",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:14",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/crossrealm",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "zrealm_crossrealm22.gno",
//             "PkgPath": "gno.land/r/crossrealm",
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
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:14](112)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:14",
//              "LastObjectSize": "255",
//     -        "ModTime": "0",
//     +        "ModTime": "47",
//              "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//              "RefCount": "1"
//          },
//     @@ -10,6 +10,11 @@
//              "T": {
//                  "@type": "/gno.RefType",
//                  "ID": "gno.land/r/demo/tests/crossrealm.FooerGetter"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "a00b77e0c471a65c4b55831f19fa15d83794fa92",
//     +            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:48"
//              }
//          }
//      }
// u[0edc46caf30c00efd87b6c272673239eafbd051e:4](7)=
//     @@ -1,10 +1,11 @@
//      {
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:4",
//     +        "IsEscaped": true,
//              "LastObjectSize": "356",
//     -        "ModTime": "0",
//     +        "ModTime": "49",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// finalizerealm["gno.land/r/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:5](5)=
//     @@ -7,14 +7,14 @@
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
//              "LastObjectSize": "239",
//     -        "ModTime": "0",
//     +        "ModTime": "15",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:4",
//              "RefCount": "1"
//          }
// finalizerealm["gno.land/r/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:5](0)=
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
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:14](1)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:14",
//              "LastObjectSize": "367",
//     -        "ModTime": "47",
//     +        "ModTime": "49",
//              "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "a00b77e0c471a65c4b55831f19fa15d83794fa92",
//     -            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:48"
//     +            "Escaped": true,
//     +            "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:7"
//              }
//          }
//      }
// u[0edc46caf30c00efd87b6c272673239eafbd051e:7](7)=
//     @@ -8,10 +8,11 @@
//          "NativePkg": "",
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:7",
//     +        "IsEscaped": true,
//              "LastObjectSize": "383",
//     -        "ModTime": "0",
//     +        "ModTime": "49",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:6",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
// u[0edc46caf30c00efd87b6c272673239eafbd051e:4](0)=
//     @@ -5,7 +5,7 @@
//              "LastObjectSize": "363",
//              "ModTime": "49",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
//     @@ -14,7 +14,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "7c146728a91cd7c6aa3141a00cede434f9601daa",
//     +            "Hash": "ecbbe489d0ec864cc480eef69ba3493a3bdbe4c6",
//                  "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:5"
//              }
//          }
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:48](-509)
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:49](-406)
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:5](0)=
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
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:50](386)={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:50",
//         "LastObjectSize": "386",
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
//                     "Line": "28"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "62",
//                     "Line": "28"
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
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:14](-1)=
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Escaped": true,
//     -            "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:7"
//     +            "Hash": "fb38b7cc96d8809dd5e5f408b9aa1a156335b6e2",
//     +            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:50"
//              }
//          }
//      }
// u[0edc46caf30c00efd87b6c272673239eafbd051e:7](0)=
//     @@ -12,7 +12,7 @@
//              "LastObjectSize": "390",
//              "ModTime": "49",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:6",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]
