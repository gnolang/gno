// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
)

func main() {
	crossing()

	f := crossrealm_b.Fooer
	cross(crossrealm.SetFooerGetter)(func() crossrealm.Fooer { return f })
	cross(crossrealm.CallFooerGetterFooSR2)()
	f.SetS("B")
	cross(crossrealm.CallFooerGetterFooSR2)()
	println(".")

	f.SetS("C")
	cross(crossrealm.SetFooerGetter)(crossrealm_b.FooerGetter)
	cross(crossrealm.CallFooerGetterFooSR2)()
	println(".")

	f.SetS("D")
	cross(crossrealm.SetFooerGetter)(crossrealm_b.FooerGetterBuilder())
	cross(crossrealm.CallFooerGetterFooSR2)()
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
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:37]={
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:37",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:36",
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
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:36]={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "be376ecae5370ffc9b531703d1dabb784f53752b",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:37"
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
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:36",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:13",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/crossrealm_test",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "35",
//             "File": "files/zrealm_crossrealm22.gno",
//             "Line": "13",
//             "PkgPath": "gno.land/r/crossrealm_test"
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
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:13]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:13",
//     -        "ModTime": "0",
//     +        "ModTime": "35",
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
//     +            "Hash": "496beec28378b765d658286561b08be89e32efee",
//     +            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:36"
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
//     +        "ModTime": "37",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
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
//     +        "ModTime": "15",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:4",
//              "RefCount": "1"
//          }
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
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
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:13]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:13",
//     -        "ModTime": "35",
//     +        "ModTime": "37",
//              "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//              "RefCount": "1"
//          },
//     @@ -12,8 +12,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "496beec28378b765d658286561b08be89e32efee",
//     -            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:36"
//     +            "Escaped": true,
//     +            "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:7"
//              }
//          }
//      }
// u[0edc46caf30c00efd87b6c272673239eafbd051e:7]=
//     @@ -8,9 +8,10 @@
//          "NativePkg": "",
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:7",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "37",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:6",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
// u[0edc46caf30c00efd87b6c272673239eafbd051e:4]=
//     @@ -4,7 +4,7 @@
//              "IsEscaped": true,
//              "ModTime": "37",
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
//     +            "Hash": "05d1e2f278c92ed1ffbe4634b915c053ac5949c2",
//                  "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:5"
//              }
//          }
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:36]
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:37]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
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
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:38]={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:38",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:13",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "Column": "62",
//             "File": "crossrealm.gno",
//             "Line": "26",
//             "PkgPath": "gno.land/r/demo/tests/crossrealm_b"
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
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:13]=
//     @@ -12,8 +12,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Escaped": true,
//     -            "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:7"
//     +            "Hash": "6a8bf31929a4f73151f3b404c376f279bacd9c81",
//     +            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:38"
//              }
//          }
//      }
// u[0edc46caf30c00efd87b6c272673239eafbd051e:7]=
//     @@ -11,7 +11,7 @@
//              "IsEscaped": true,
//              "ModTime": "37",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:6",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/crossrealm_test"]
