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
	f.SetS(cross(cur), "B")
	crossrealm.CallFooerGetterFooCross(cross(cur))
	println(".")

	f.SetS(cross(cur), "C")
	crossrealm.SetFooerGetter(cross(cur), crossrealm_b.FooerGetter)
	crossrealm.CallFooerGetterFooCross(cross(cur))
	println(".")

	f.SetS(cross(cur), "D")
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
// c[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5](400)={
//     "ObjectInfo": {
//         "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5",
//         "LastObjectSize": "400",
//         "ModTime": "0",
//         "OwnerID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:47",
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
// c[07e973c6aeab1f1cb747d0903cb1e4d369226376:47](508)={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "917738a2d5fd1153772f22cbb83cd2848aaa79d8",
//                 "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5"
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
//         "ID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:47",
//         "LastObjectSize": "508",
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
//                 "PkgPath": "",
//                 "Tag": "",
//                 "Type": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests/vm/crossrealm.Fooer"
//                 }
//             }
//         ]
//     }
// }
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:14](112)=
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
//     +            "Hash": "b9aed7472a5b54f31f0ee9b1f7d88e49588d1744",
//     +            "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:47"
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
//     +        "ModTime": "47",
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
//     +        "ModTime": "47",
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
//     @@ -4,7 +4,7 @@
//              "ID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:2",
//              "IsEscaped": true,
//              "LastObjectSize": "8458",
//     -        "ModTime": "46",
//     +        "ModTime": "47",
//              "RefCount": "3"
//          },
//          "Parent": null,
//     @@ -276,7 +276,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "68684cec1e3a40b9022f09802324c8388b874458",
//     +                "Hash": "bfdd387796936c99a85ea1b63ba921760e47638d",
//                      "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14"
//                  }
//              },
// u[06279d1e03ecb38b84822f9caf584bd16913470d:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "IsEscaped": true,
//              "LastObjectSize": "2495",
//     -        "ModTime": "0",
//     +        "ModTime": "47",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e6301dc45d62d2b7659c3f5fff00f83b0a69f7d7",
//     +                "Hash": "292fbb0abe36bcf5772f1de75895537cfeda1302",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:3"
//                  }
//              },
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
//     +        "ModTime": "20",
//              "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "RefCount": "1"
//          }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -3,7 +3,7 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "IsEscaped": true,
//              "LastObjectSize": "361",
//     -        "ModTime": "47",
//     +        "ModTime": "20",
//              "RefCount": "2"
//          },
//          "Value": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "534961c4c33e00090c5b38e73cdddaa534514e00",
//     +            "Hash": "159cc6ae11c538607def2c6232ce144dd2675892",
//                  "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
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
//     -            "Hash": "159cc6ae11c538607def2c6232ce144dd2675892",
//     +            "Hash": "62090f3fe721220b525ff7356320d0fb8328573f",
//                  "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// c[07e973c6aeab1f1cb747d0903cb1e4d369226376:48](381)={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": true,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:48",
//         "LastObjectSize": "381",
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
//                     "Column": "63",
//                     "Line": "29"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "23",
//                     "Line": "29"
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
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests/vm/crossrealm.Fooer"
//                 }
//             }
//         ]
//     }
// }
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:14](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14",
//              "LastObjectSize": "365",
//     -        "ModTime": "46",
//     +        "ModTime": "47",
//              "OwnerID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:2",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "b9aed7472a5b54f31f0ee9b1f7d88e49588d1744",
//     -            "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:47"
//     +            "Hash": "fd326a185b324ec7433d08d6cc7aa1a9424ce573",
//     +            "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:48"
//              }
//          }
//      }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -3,8 +3,8 @@
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "IsEscaped": true,
//              "LastObjectSize": "361",
//     -        "ModTime": "20",
//     -        "RefCount": "2"
//     +        "ModTime": "48",
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:2",
//              "IsEscaped": true,
//              "LastObjectSize": "8458",
//     -        "ModTime": "47",
//     +        "ModTime": "48",
//              "RefCount": "3"
//          },
//          "Parent": null,
//     @@ -276,7 +276,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "bfdd387796936c99a85ea1b63ba921760e47638d",
//     +                "Hash": "d2e2b6ef8a821e22758a1922d6bb26fd63779f8e",
//                      "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14"
//                  }
//              },
// d[07e973c6aeab1f1cb747d0903cb1e4d369226376:47](-508)
// d[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:5](-400)
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
//     -        "ModTime": "48",
//     +        "ModTime": "20",
//              "RefCount": "1"
//          },
//          "Value": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "62090f3fe721220b525ff7356320d0fb8328573f",
//     +            "Hash": "238d23e0a421711881a6a1f3d955be12c5e1edd6",
//                  "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// c[07e973c6aeab1f1cb747d0903cb1e4d369226376:49](382)={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": true,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:49",
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
//                     "Line": "30"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "62",
//                     "Line": "30"
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
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/tests/vm/crossrealm.Fooer"
//                 }
//             }
//         ]
//     }
// }
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:14](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14",
//              "LastObjectSize": "365",
//     -        "ModTime": "47",
//     +        "ModTime": "48",
//              "OwnerID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:2",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "fd326a185b324ec7433d08d6cc7aa1a9424ce573",
//     -            "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:48"
//     +            "Hash": "14cfd1e53f2f6ead1a92fccdf7e6028ee4dccb1b",
//     +            "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:49"
//              }
//          }
//      }
// u[07e973c6aeab1f1cb747d0903cb1e4d369226376:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:2",
//              "IsEscaped": true,
//              "LastObjectSize": "8458",
//     -        "ModTime": "48",
//     +        "ModTime": "49",
//              "RefCount": "3"
//          },
//          "Parent": null,
//     @@ -276,7 +276,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "d2e2b6ef8a821e22758a1922d6bb26fd63779f8e",
//     +                "Hash": "8665dd09ae1244a57772e4f549ff86a99edb9c10",
//                      "ObjectID": "07e973c6aeab1f1cb747d0903cb1e4d369226376:14"
//                  }
//              },
// d[07e973c6aeab1f1cb747d0903cb1e4d369226376:48](-381)
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]
