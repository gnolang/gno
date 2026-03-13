// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm"
	"gno.land/r/tests/vm/crossrealm_b"
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
// hello A cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// hello B cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .
// hello C cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .
// hello D cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// c[a7e973c6aeab1f1cb747d0903cb1e4d369226376:45](404)={
//     "ObjectInfo": {
//         "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:45",
//         "LastObjectSize": "404",
//         "ModTime": "0",
//         "OwnerID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:44",
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
//                 "Escaped": true,
//                 "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:4"
//             },
//             "Index": "0",
//             "TV": null
//         }
//     }
// }
// c[a7e973c6aeab1f1cb747d0903cb1e4d369226376:44](507)={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "ceee62271c4bf99abb62ad71f525a8bde9ad6412",
//                 "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:45"
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
//         "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:44",
//         "LastObjectSize": "507",
//         "ModTime": "0",
//         "OwnerID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:14",
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
//                     "ID": "gno.land/r/tests/vm/crossrealm.Fooer"
//                 }
//             }
//         ]
//     }
// }
// u[a7e973c6aeab1f1cb747d0903cb1e4d369226376:14](112)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:14",
//              "LastObjectSize": "253",
//     -        "ModTime": "0",
//     +        "ModTime": "43",
//              "OwnerID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:2",
//              "RefCount": "1"
//          },
//     @@ -10,6 +10,11 @@
//              "T": {
//                  "@type": "/gno.RefType",
//                  "ID": "gno.land/r/tests/vm/crossrealm.FooerGetter"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "52c5a1ccabf063e4d43938fc761b915e74d4f10a",
//     +            "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:44"
//              }
//          }
//      }
// u[46279d1e03ecb38b84822f9caf584bd16913470d:4](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:4",
//     +        "IsEscaped": true,
//              "LastObjectSize": "354",
//     -        "ModTime": "0",
//     -        "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:3",
//     -        "RefCount": "1"
//     +        "ModTime": "45",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[46279d1e03ecb38b84822f9caf584bd16913470d:3](8)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:3",
//              "LastObjectSize": "399",
//     -        "ModTime": "0",
//     +        "ModTime": "45",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "3ee94e5f0fd68b290b853e345faf220a8b7ffc81",
//     +                "Escaped": true,
//                      "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:4"
//                  },
//                  "Index": "0",
// u[a7e973c6aeab1f1cb747d0903cb1e4d369226376:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:2",
//              "IsEscaped": true,
//              "LastObjectSize": "8731",
//     -        "ModTime": "43",
//     +        "ModTime": "45",
//              "RefCount": "3"
//          },
//          "Parent": null,
//     @@ -482,7 +482,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e38cd85c59431f6180aff163737c2b3822b89336",
//     +                "Hash": "59c676a3858135eda39f76d35f661e43b495262a",
//                      "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:14"
//                  }
//              },
// u[46279d1e03ecb38b84822f9caf584bd16913470d:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "IsEscaped": true,
//              "LastObjectSize": "3910",
//     -        "ModTime": "0",
//     +        "ModTime": "45",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -340,7 +340,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "c1295ed2e4e057ac5574254bd7f4916298301a79",
//     +                "Hash": "7355ab81ed22fa225b2e3019a8700f0ac8fdec82",
//                      "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:3"
//                  }
//              },
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[46279d1e03ecb38b84822f9caf584bd16913470d:5](5)=
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
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:5",
//              "LastObjectSize": "239",
//     -        "ModTime": "0",
//     +        "ModTime": "18",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "RefCount": "1"
//          }
// u[46279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -3,7 +3,7 @@
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "IsEscaped": true,
//              "LastObjectSize": "361",
//     -        "ModTime": "45",
//     +        "ModTime": "18",
//              "RefCount": "2"
//          },
//          "Value": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "23c79341f8583900b8e4138982dda98181fc2d8c",
//     +            "Hash": "86833e4800b692d909ab6c1539db7a7731aec8ca",
//                  "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[46279d1e03ecb38b84822f9caf584bd16913470d:5](0)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "B"
//     +                "value": "C"
//                  }
//              }
//          ],
// u[46279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "86833e4800b692d909ab6c1539db7a7731aec8ca",
//     +            "Hash": "6e0f68a9de37b3e945fe297f093b079c869a3a07",
//                  "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// u[a7e973c6aeab1f1cb747d0903cb1e4d369226376:14](1)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:14",
//              "LastObjectSize": "365",
//     -        "ModTime": "43",
//     +        "ModTime": "45",
//              "OwnerID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:2",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "52c5a1ccabf063e4d43938fc761b915e74d4f10a",
//     -            "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:44"
//     +            "Escaped": true,
//     +            "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:7"
//              }
//          }
//      }
// u[46279d1e03ecb38b84822f9caf584bd16913470d:7](7)=
//     @@ -8,10 +8,10 @@
//          "NativePkg": "",
//          "ObjectInfo": {
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:7",
//     +        "IsEscaped": true,
//              "LastObjectSize": "377",
//     -        "ModTime": "0",
//     -        "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:6",
//     -        "RefCount": "1"
//     +        "ModTime": "45",
//     +        "RefCount": "2"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/tests/vm/crossrealm_b",
// u[46279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -3,8 +3,8 @@
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "IsEscaped": true,
//              "LastObjectSize": "361",
//     -        "ModTime": "18",
//     -        "RefCount": "2"
//     +        "ModTime": "45",
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
// u[46279d1e03ecb38b84822f9caf584bd16913470d:6](7)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:6",
//              "LastObjectSize": "381",
//     -        "ModTime": "0",
//     +        "ModTime": "45",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "RefCount": "1"
//          },
//     @@ -24,7 +24,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "3892bee1fa60826be1245fceca81ea6e0da2efb5",
//     +            "Escaped": true,
//                  "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:7"
//              }
//          }
// u[a7e973c6aeab1f1cb747d0903cb1e4d369226376:2](0)=
//     @@ -482,7 +482,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "59c676a3858135eda39f76d35f661e43b495262a",
//     +                "Hash": "389752abf0153407b24be9602c6ff54c8c3768d2",
//                      "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:14"
//                  }
//              },
// u[46279d1e03ecb38b84822f9caf584bd16913470d:2](0)=
//     @@ -350,7 +350,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "4cce0616c82133d94ec1b2ab8d5c2741ca6c786d",
//     +                "Hash": "b5fbc65fa06e72495f5e1844133fc23b8f96f55f",
//                      "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:6"
//                  }
//              },
// d[a7e973c6aeab1f1cb747d0903cb1e4d369226376:44](-507)
// d[a7e973c6aeab1f1cb747d0903cb1e4d369226376:45](-404)
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[46279d1e03ecb38b84822f9caf584bd16913470d:5](0)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "C"
//     +                "value": "D"
//                  }
//              }
//          ],
// u[46279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -3,7 +3,7 @@
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "IsEscaped": true,
//              "LastObjectSize": "361",
//     -        "ModTime": "45",
//     +        "ModTime": "18",
//              "RefCount": "1"
//          },
//          "Value": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "6e0f68a9de37b3e945fe297f093b079c869a3a07",
//     +            "Hash": "9784831f03fd643b16a793883de3a56725d0126c",
//                  "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// c[a7e973c6aeab1f1cb747d0903cb1e4d369226376:46](380)={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:46",
//         "LastObjectSize": "380",
//         "ModTime": "0",
//         "OwnerID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:14",
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
// u[a7e973c6aeab1f1cb747d0903cb1e4d369226376:14](-1)=
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Escaped": true,
//     -            "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:7"
//     +            "Hash": "cf4b669212d7a060336ace4ecf29fb8c45b0a195",
//     +            "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:46"
//              }
//          }
//      }
// u[46279d1e03ecb38b84822f9caf584bd16913470d:7](0)=
//     @@ -11,7 +11,7 @@
//              "IsEscaped": true,
//              "LastObjectSize": "384",
//              "ModTime": "45",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/tests/vm/crossrealm_b",
// u[a7e973c6aeab1f1cb747d0903cb1e4d369226376:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:2",
//              "IsEscaped": true,
//              "LastObjectSize": "8731",
//     -        "ModTime": "45",
//     +        "ModTime": "46",
//              "RefCount": "3"
//          },
//          "Parent": null,
//     @@ -482,7 +482,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "389752abf0153407b24be9602c6ff54c8c3768d2",
//     +                "Hash": "7231ea60e1ca37243cea71c0ac4eb6984eabfcf3",
//                      "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:14"
//                  }
//              },
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]
