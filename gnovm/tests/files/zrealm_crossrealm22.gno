// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm"
	"gno.land/r/tests/vm/crossrealm_b"
)

func main(cur realm) {
	f := crossrealm_b.Fooer
	crossrealm.SetFooerGetter(cross2(cur), func() crossrealm.Fooer { return f })
	crossrealm.CallFooerGetterFooCross(cross2(cur))
	f.SetS("B")
	crossrealm.CallFooerGetterFooCross(cross2(cur))
	println(".")

	f.SetS("C")
	crossrealm.SetFooerGetter(cross2(cur), crossrealm_b.FooerGetter)
	crossrealm.CallFooerGetterFooCross(cross2(cur))
	println(".")

	f.SetS("D")
	crossrealm.SetFooerGetter(cross2(cur), crossrealm_b.FooerGetterBuilder())
	crossrealm.CallFooerGetterFooCross(cross2(cur))
	println(".")
}

// Output:
// hello A cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// hello B cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .
// hello C cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .
// hello D cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// c[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:6](399)={
//     "ObjectInfo": {
//         "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:6",
//         "LastObjectSize": "399",
//         "ModTime": "0",
//         "OwnerID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5",
//         "RefCount": "1"
//     },
//     "Value": {
//         "N": "UmVhRG9OTFk=",
//         "T": {
//             "@type": "/gno.PointerType",
//             "Elt": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/tests/vm/crossrealm_b.fooer"
//             }
//         },
//         "V": {
//             "@type": "/gno.PointerValue",
//             "Base": {
//                 "@type": "/gno.RefValue",
//                 "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:4"
//             },
//             "Index": "0",
//             "TV": null
//         }
//     }
// }
// c[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5](505)={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "c17906071fc752254f15f17ff49db280d68fb67d",
//                 "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:6"
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
//         "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5",
//         "LastObjectSize": "505",
//         "ModTime": "0",
//         "OwnerID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14",
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
//                     "Column": "77",
//                     "Line": "11"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "41",
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
//                     "ID": "gno.land/r/tests/vm/crossrealm.Fooer"
//                 }
//             }
//         ]
//     }
// }
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:14](111)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14",
//              "LastObjectSize": "253",
//     -        "ModTime": "0",
//     +        "ModTime": "43",
//              "OwnerID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:2",
//              "RefCount": "1"
//          },
//     @@ -10,6 +10,11 @@
//              "T": {
//                  "@type": "/gno.RefType",
//                  "ID": "gno.land/r/tests/vm/crossrealm.FooerGetter"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "a9137b0c1ba2fbe0a4c917c8e5674cb6dad241df",
//     +            "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5"
//              }
//          }
//      }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:4](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//     +        "IsEscaped": true,
//              "LastObjectSize": "354",
//     -        "ModTime": "0",
//     -        "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:3",
//     -        "RefCount": "1"
//     +        "ModTime": "43",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[06279d1e03ecb38b84822f9caf584bd16913470d:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:3",
//              "LastObjectSize": "399",
//     -        "ModTime": "0",
//     +        "ModTime": "43",
//              "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,6 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "2cf29cfddfad3bf584a0b0607697df3d23ecf04c",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:4"
//                  },
//                  "Index": "0",
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:2](0)=
//     @@ -268,7 +268,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "68684cec1e3a40b9022f09802324c8388b874458",
//     +                "Hash": "b0da0b66af78ca78fe51ddecbac80a86372bcc92",
//                      "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14"
//                  }
//              },
// u[06279d1e03ecb38b84822f9caf584bd16913470d:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "IsEscaped": true,
//              "LastObjectSize": "2103",
//     -        "ModTime": "0",
//     +        "ModTime": "43",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e6301dc45d62d2b7659c3f5fff00f83b0a69f7d7",
//     +                "Hash": "58b6ef7fbd90cca4b6157fb97f9017b4bb55d0d2",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:3"
//                  }
//              },
// finalizerealm["gno.land/r/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[06279d1e03ecb38b84822f9caf584bd16913470d:5](5)=
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
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:5",
//              "LastObjectSize": "239",
//     -        "ModTime": "0",
//     +        "ModTime": "18",
//              "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "RefCount": "1"
//          }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -3,7 +3,7 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "IsEscaped": true,
//              "LastObjectSize": "361",
//     -        "ModTime": "43",
//     +        "ModTime": "18",
//              "RefCount": "2"
//          },
//          "Value": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "534961c4c33e00090c5b38e73cdddaa534514e00",
//     +            "Hash": "cf8abd9e6c10a053c7eccb921d67b74c5505729c",
//                  "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
// finalizerealm["gno.land/r/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[06279d1e03ecb38b84822f9caf584bd16913470d:5](0)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "B"
//     +                "value": "C"
//                  }
//              }
//          ],
// u[06279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "cf8abd9e6c10a053c7eccb921d67b74c5505729c",
//     +            "Hash": "5aef97fff6d6cee13dcfc7c35cd6eead9c972ebd",
//                  "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:14](0)=
//     @@ -13,8 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "a9137b0c1ba2fbe0a4c917c8e5674cb6dad241df",
//     -            "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5"
//     +            "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:7"
//              }
//          }
//      }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:7](7)=
//     @@ -8,10 +8,10 @@
//          "NativePkg": "",
//          "ObjectInfo": {
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:7",
//     +        "IsEscaped": true,
//              "LastObjectSize": "377",
//     -        "ModTime": "0",
//     -        "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:6",
//     -        "RefCount": "1"
//     +        "ModTime": "43",
//     +        "RefCount": "2"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/tests/vm/crossrealm_b",
// u[06279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -3,8 +3,8 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "IsEscaped": true,
//              "LastObjectSize": "361",
//     -        "ModTime": "18",
//     -        "RefCount": "2"
//     +        "ModTime": "43",
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
// u[06279d1e03ecb38b84822f9caf584bd16913470d:6](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:6",
//              "LastObjectSize": "381",
//     -        "ModTime": "0",
//     +        "ModTime": "43",
//              "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "RefCount": "1"
//          },
//     @@ -24,7 +24,6 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "ce24e8f1fb0aa702f5f718aca53a21192a560d4a",
//                  "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:7"
//              }
//          }
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:2](0)=
//     @@ -268,7 +268,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "b0da0b66af78ca78fe51ddecbac80a86372bcc92",
//     +                "Hash": "97020ce5ecf9791c98331a8bba382c5a55d0cd71",
//                      "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14"
//                  }
//              },
// u[06279d1e03ecb38b84822f9caf584bd16913470d:2](0)=
//     @@ -56,7 +56,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5e8e8266d126ede07ea788abc2e47ba77bd58746",
//     +                "Hash": "b82a3655543a4fc6f3e053ff0931cae6ccaca38a",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:6"
//                  }
//              },
// d[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5](-505)
// d[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:6](-399)
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[06279d1e03ecb38b84822f9caf584bd16913470d:5](0)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "C"
//     +                "value": "D"
//                  }
//              }
//          ],
// u[06279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -3,7 +3,7 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "IsEscaped": true,
//              "LastObjectSize": "361",
//     -        "ModTime": "43",
//     +        "ModTime": "18",
//              "RefCount": "1"
//          },
//          "Value": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "5aef97fff6d6cee13dcfc7c35cd6eead9c972ebd",
//     +            "Hash": "4eb698d9935d15626f4a5ea91a4cd2b7d374a93c",
//                  "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// c[06279d1e03ecb38b84822f9caf584bd16913470d:19](380)={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:19",
//         "LastObjectSize": "380",
//         "ModTime": "0",
//         "OwnerID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/tests/vm/crossrealm_b",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "crossrealm.gno",
//             "PkgPath": "gno.land/r/tests/vm/crossrealm_b",
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
//                     "ID": "gno.land/r/tests/vm/crossrealm.Fooer"
//                 }
//             }
//         ]
//     }
// }
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:14](1)=
//     @@ -13,7 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:7"
//     +            "Hash": "5a5cc8020679dadb537ed1f992d13e5157c088d3",
//     +            "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:19"
//              }
//          }
//      }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:7](0)=
//     @@ -11,7 +11,7 @@
//              "IsEscaped": true,
//              "LastObjectSize": "384",
//              "ModTime": "43",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/tests/vm/crossrealm_b",
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:2](0)=
//     @@ -268,7 +268,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "97020ce5ecf9791c98331a8bba382c5a55d0cd71",
//     +                "Hash": "db04f18876eb212d96492749250b88d097a1095d",
//                      "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14"
//                  }
//              },
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]
