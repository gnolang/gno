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
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:26]={
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:26",
//         "ModTime": "0",
//         "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:25",
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
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:25]={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "0f1d56943a51eee9cf77c2dabfbcd09a9fce59f5",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:26"
//             }
//         }
//     ],
//     "FileName": "",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:25",
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
//             "Column": "40",
//             "File": "files/zrealm_crossrealm22.gno",
//             "Line": "13",
//             "PkgPath": "gno.land/r/crossrealm_test"
//         }
//     },
//     "SwitchRealm": false,
//     "Type": {
//         "@type": "/gno.FuncType",
//         "IsClosure": false,
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
//     +        "ModTime": "24",
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
//     +            "Hash": "fc865d3e354500ec5fbefdd6e008d12b10d6e95c",
//     +            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:25"
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
//     +        "ModTime": "26",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
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
//     +        "ModTime": "14",
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
// u[1712ac7adcfdc8e58a67e5615e20fb312394c4df:13]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:13",
//     -        "ModTime": "24",
//     +        "ModTime": "26",
//              "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
//              "RefCount": "1"
//          },
//     @@ -12,8 +12,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "fc865d3e354500ec5fbefdd6e008d12b10d6e95c",
//     -            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:25"
//     +            "Escaped": true,
//     +            "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:7"
//              }
//          }
//      }
// u[0edc46caf30c00efd87b6c272673239eafbd051e:7]=
//     @@ -7,9 +7,10 @@
//          "NativePkg": "",
//          "ObjectInfo": {
//              "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:7",
//     -        "ModTime": "0",
//     +        "IsEscaped": true,
//     +        "ModTime": "26",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:6",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
// u[0edc46caf30c00efd87b6c272673239eafbd051e:4]=
//     @@ -4,7 +4,7 @@
//              "IsEscaped": true,
//              "ModTime": "26",
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
//     +            "Hash": "631b3a7bdaa9fa373463bf446868e762eb81692f",
//                  "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:5"
//              }
//          }
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:25]
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:26]
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
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:27]={
//     "FileName": "",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:27",
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
//     "SwitchRealm": false,
//     "Type": {
//         "@type": "/gno.FuncType",
//         "IsClosure": false,
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
//     +            "Hash": "ab4ad7390e55495fe0f4afcfe582b46be527d027",
//     +            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:27"
//              }
//          }
//      }
// u[0edc46caf30c00efd87b6c272673239eafbd051e:7]=
//     @@ -10,7 +10,7 @@
//              "IsEscaped": true,
//              "ModTime": "26",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:6",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]
