// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/tests/crossrealm"
	"gno.land/r/tests/tests/crossrealm_b"
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
// hello A cur=gno.land/r/tests/tests/crossrealm_b prev=gno.land/r/tests/tests/crossrealm
// hello B cur=gno.land/r/tests/tests/crossrealm_b prev=gno.land/r/tests/tests/crossrealm
// .
// hello C cur=gno.land/r/tests/tests/crossrealm_b prev=gno.land/r/tests/tests/crossrealm
// .
// hello D cur=gno.land/r/tests/tests/crossrealm_b prev=gno.land/r/tests/tests/crossrealm
// .

// Realm:
// finalizerealm["gno.land/r/tests/tests/crossrealm"]
// c[83c395a7172ae8aebc131bd64cc048bbd2b31851:45](407)={
//     "ObjectInfo": {
//         "ID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:45",
//         "LastObjectSize": "407",
//         "ModTime": "0",
//         "OwnerID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:44",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.PointerType",
//             "Elt": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/tests/tests/crossrealm_b.fooer"
//             }
//         },
//         "V": {
//             "@type": "/gno.PointerValue",
//             "Base": {
//                 "@type": "/gno.RefValue",
//                 "Escaped": true,
//                 "ObjectID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:4"
//             },
//             "Index": "0",
//             "TV": null
//         }
//     }
// }
// c[83c395a7172ae8aebc131bd64cc048bbd2b31851:44](510)={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "b21569a3d78e9993a29aa1dbeb462600e10fc19d",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:45"
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
//         "ID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:44",
//         "LastObjectSize": "510",
//         "ModTime": "0",
//         "OwnerID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:14",
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
//                     "ID": "gno.land/r/tests/tests/crossrealm.Fooer"
//                 }
//             }
//         ]
//     }
// }
// u[83c395a7172ae8aebc131bd64cc048bbd2b31851:14](112)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:14",
//              "LastObjectSize": "256",
//     -        "ModTime": "0",
//     +        "ModTime": "43",
//              "OwnerID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:2",
//              "RefCount": "1"
//          },
//     @@ -10,6 +10,11 @@
//              "T": {
//                  "@type": "/gno.RefType",
//                  "ID": "gno.land/r/tests/tests/crossrealm.FooerGetter"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "19a891527c775a06c4e2e7315beba0aea71e30e9",
//     +            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:44"
//              }
//          }
//      }
// u[3003d3e51bf2f146cfba386aa7161085b561ec2d:4](7)=
//     @@ -1,10 +1,11 @@
//      {
//          "ObjectInfo": {
//              "ID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:4",
//     +        "IsEscaped": true,
//              "LastObjectSize": "357",
//     -        "ModTime": "0",
//     +        "ModTime": "45",
//              "OwnerID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:3",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// finalizerealm["gno.land/r/tests/tests/crossrealm_b"]
// finalizerealm["gno.land/r/tests/tests/crossrealm"]
// finalizerealm["gno.land/r/tests/tests/crossrealm_b"]
// u[3003d3e51bf2f146cfba386aa7161085b561ec2d:5](5)=
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
//              "ID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:5",
//              "LastObjectSize": "239",
//     -        "ModTime": "0",
//     +        "ModTime": "15",
//              "OwnerID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:4",
//              "RefCount": "1"
//          }
// finalizerealm["gno.land/r/tests/tests/crossrealm_b"]
// finalizerealm["gno.land/r/tests/tests/crossrealm"]
// finalizerealm["gno.land/r/tests/tests/crossrealm_b"]
// u[3003d3e51bf2f146cfba386aa7161085b561ec2d:5](0)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "B"
//     +                "value": "C"
//                  }
//              }
//          ],
// finalizerealm["gno.land/r/tests/tests/crossrealm"]
// u[83c395a7172ae8aebc131bd64cc048bbd2b31851:14](1)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:14",
//              "LastObjectSize": "368",
//     -        "ModTime": "43",
//     +        "ModTime": "45",
//              "OwnerID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:2",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "19a891527c775a06c4e2e7315beba0aea71e30e9",
//     -            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:44"
//     +            "Escaped": true,
//     +            "ObjectID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:7"
//              }
//          }
//      }
// u[3003d3e51bf2f146cfba386aa7161085b561ec2d:7](7)=
//     @@ -8,10 +8,11 @@
//          "NativePkg": "",
//          "ObjectInfo": {
//              "ID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:7",
//     +        "IsEscaped": true,
//              "LastObjectSize": "386",
//     -        "ModTime": "0",
//     +        "ModTime": "45",
//              "OwnerID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:6",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/tests/tests/crossrealm_b",
// u[3003d3e51bf2f146cfba386aa7161085b561ec2d:4](0)=
//     @@ -5,7 +5,7 @@
//              "LastObjectSize": "364",
//              "ModTime": "45",
//              "OwnerID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:3",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
//     @@ -14,7 +14,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "6fd7ebfcd0552e852c342a2652b9281c11e192d2",
//     +            "Hash": "0329cfcf74f963d00e2d74ec484fd94831966ef6",
//                  "ObjectID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:5"
//              }
//          }
// d[83c395a7172ae8aebc131bd64cc048bbd2b31851:44](-510)
// d[83c395a7172ae8aebc131bd64cc048bbd2b31851:45](-407)
// finalizerealm["gno.land/r/tests/tests/crossrealm_b"]
// finalizerealm["gno.land/r/tests/tests/crossrealm"]
// finalizerealm["gno.land/r/tests/tests/crossrealm_b"]
// u[3003d3e51bf2f146cfba386aa7161085b561ec2d:5](0)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "C"
//     +                "value": "D"
//                  }
//              }
//          ],
// finalizerealm["gno.land/r/tests/tests/crossrealm"]
// c[83c395a7172ae8aebc131bd64cc048bbd2b31851:46](389)={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:46",
//         "LastObjectSize": "389",
//         "ModTime": "0",
//         "OwnerID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:14",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/tests/tests/crossrealm_b",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "crossrealm.gno",
//             "PkgPath": "gno.land/r/tests/tests/crossrealm_b",
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
//                     "ID": "gno.land/r/tests/tests/crossrealm.Fooer"
//                 }
//             }
//         ]
//     }
// }
// u[83c395a7172ae8aebc131bd64cc048bbd2b31851:14](-1)=
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Escaped": true,
//     -            "ObjectID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:7"
//     +            "Hash": "b34c145de5eaa0ce2049e351b798bff9726cd88d",
//     +            "ObjectID": "83c395a7172ae8aebc131bd64cc048bbd2b31851:46"
//              }
//          }
//      }
// u[3003d3e51bf2f146cfba386aa7161085b561ec2d:7](0)=
//     @@ -12,7 +12,7 @@
//              "LastObjectSize": "393",
//              "ModTime": "45",
//              "OwnerID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:6",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/tests/tests/crossrealm_b",
// finalizerealm["gno.land/r/tests/tests/crossrealm_b"]
// finalizerealm["gno.land/r/tests/tests/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]
