// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm"
	"gno.land/r/tests/vm/crossrealm_b"
)

func main(cur realm) {
	f := crossrealm_b.Fooer
	crossrealm.SetFooerGetter(cross(cur), func() crossrealm.Fooer { return f })
	crossrealm.CallFooerGetterFooCross(cross(cur))
	f.SetS("B")
	crossrealm.CallFooerGetterFooCross(cross(cur))
	println(".")

	f.SetS("C")
	crossrealm.SetFooerGetter(cross(cur), crossrealm_b.FooerGetter)
	crossrealm.CallFooerGetterFooCross(cross(cur))
	println(".")

	f.SetS("D")
	crossrealm.SetFooerGetter(cross(cur), crossrealm_b.FooerGetterBuilder())
	crossrealm.CallFooerGetterFooCross(cross(cur))
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
// c[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5](507)={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "6e22577dee90926a3bc184c1cd0b706b341aff5d",
//                 "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:6"
//             }
//         }
//     ],
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": true,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5",
//         "LastObjectSize": "507",
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
//                     "Column": "76",
//                     "Line": "11"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "40",
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
//     +        "ModTime": "46",
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
//     +            "Hash": "8d8d45622e3a5236372a1d9b560ef5a3a1c29634",
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
//     +        "ModTime": "46",
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
//     +        "ModTime": "46",
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
//     +                "Hash": "5cab66f9e9dcdb1c5ff2afb9b0bca8c3b9bad879",
//                      "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14"
//                  }
//              },
// u[06279d1e03ecb38b84822f9caf584bd16913470d:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "IsEscaped": true,
//              "LastObjectSize": "2317",
//     -        "ModTime": "0",
//     +        "ModTime": "46",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e6301dc45d62d2b7659c3f5fff00f83b0a69f7d7",
//     +                "Hash": "ee6677faa9daebe7962dd190f3c30fea3e832866",
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
//     +        "ModTime": "19",
//              "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "RefCount": "1"
//          }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -3,7 +3,7 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "IsEscaped": true,
//              "LastObjectSize": "361",
//     -        "ModTime": "46",
//     +        "ModTime": "19",
//              "RefCount": "2"
//          },
//          "Value": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "534961c4c33e00090c5b38e73cdddaa534514e00",
//     +            "Hash": "94cd03d7e1847141ea0a15de76fb02421ea9f7a3",
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
//     -            "Hash": "94cd03d7e1847141ea0a15de76fb02421ea9f7a3",
//     +            "Hash": "05e62928ec6ae2d2ba062f805517e6332a2bfc1c",
//                  "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:14](0)=
//     @@ -13,8 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "8d8d45622e3a5236372a1d9b560ef5a3a1c29634",
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
//              "LastObjectSize": "379",
//     -        "ModTime": "0",
//     -        "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:6",
//     -        "RefCount": "1"
//     +        "ModTime": "46",
//     +        "RefCount": "2"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/tests/vm/crossrealm_b",
// u[06279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -3,8 +3,8 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "IsEscaped": true,
//              "LastObjectSize": "361",
//     -        "ModTime": "19",
//     -        "RefCount": "2"
//     +        "ModTime": "46",
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
//     +        "ModTime": "46",
//              "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "RefCount": "1"
//          },
//     @@ -24,7 +24,6 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "6eec283cfa79c57948d2ab5646d75f34c461da56",
//                  "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:7"
//              }
//          }
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:2](0)=
//     @@ -268,7 +268,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5cab66f9e9dcdb1c5ff2afb9b0bca8c3b9bad879",
//     +                "Hash": "f304224c3678ae566effa9f679429b3552eabcff",
//                      "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14"
//                  }
//              },
// u[06279d1e03ecb38b84822f9caf584bd16913470d:2](0)=
//     @@ -56,7 +56,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e01c8f7f59da8b879785967f4d51ffb600a538b0",
//     +                "Hash": "c3504e35dfe4fe5650b98d15fc520745d1714f75",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:6"
//                  }
//              },
// d[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5](-507)
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
//     -        "ModTime": "46",
//     +        "ModTime": "19",
//              "RefCount": "1"
//          },
//          "Value": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "05e62928ec6ae2d2ba062f805517e6332a2bfc1c",
//     +            "Hash": "1735db3b3cdf4b712a999a650368ba8492104c5b",
//                  "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// c[06279d1e03ecb38b84822f9caf584bd16913470d:20](382)={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": true,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:20",
//         "LastObjectSize": "382",
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
//     +            "Hash": "ec75abe8cd08aa82fabeeb72f7457f9d6aa2f91c",
//     +            "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:20"
//              }
//          }
//      }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:7](0)=
//     @@ -11,7 +11,7 @@
//              "IsEscaped": true,
//              "LastObjectSize": "386",
//              "ModTime": "46",
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
//     -                "Hash": "f304224c3678ae566effa9f679429b3552eabcff",
//     +                "Hash": "60191778a1764445a69e67328f4f685013339443",
//                      "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14"
//                  }
//              },
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]
