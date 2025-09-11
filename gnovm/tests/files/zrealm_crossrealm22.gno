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

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// c[a7e973c6aeab1f1cb747d0903cb1e4d369226376:55](404)={
//     "ObjectInfo": {
//         "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:55",
//         "LastObjectSize": "404",
//         "ModTime": "0",
//         "OwnerID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:54",
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
//                 "Escaped": true,
//                 "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:4"
//             },
//             "Index": "0",
//             "TV": null
//         }
//     }
// }
// c[a7e973c6aeab1f1cb747d0903cb1e4d369226376:54](507)={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "a3b6555e3e24e0096ad1fb65412b3bf5cd6eaac2",
//                 "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:55"
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
//         "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:54",
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
//     +        "ModTime": "53",
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
//     +            "Hash": "a36227cc1e34e9c32a4ac12012be18631eaee7f1",
//     +            "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:54"
//              }
//          }
//      }
// u[46279d1e03ecb38b84822f9caf584bd16913470d:4](7)=
//     @@ -1,10 +1,11 @@
//      {
//          "ObjectInfo": {
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:4",
//     +        "IsEscaped": true,
//              "LastObjectSize": "354",
//     -        "ModTime": "0",
//     +        "ModTime": "55",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:3",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// finalizerealm["gno.land/r/crossrealm"]
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
//     +        "ModTime": "15",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "RefCount": "1"
//          }
// finalizerealm["gno.land/r/crossrealm"]
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
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// u[a7e973c6aeab1f1cb747d0903cb1e4d369226376:14](1)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:14",
//              "LastObjectSize": "365",
//     -        "ModTime": "53",
//     +        "ModTime": "55",
//              "OwnerID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:2",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "a36227cc1e34e9c32a4ac12012be18631eaee7f1",
//     -            "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:54"
//     +            "Escaped": true,
//     +            "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:7"
//              }
//          }
//      }
// u[46279d1e03ecb38b84822f9caf584bd16913470d:7](7)=
//     @@ -8,10 +8,11 @@
//          "NativePkg": "",
//          "ObjectInfo": {
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:7",
//     +        "IsEscaped": true,
//              "LastObjectSize": "377",
//     -        "ModTime": "0",
//     +        "ModTime": "55",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:6",
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/tests/vm/crossrealm_b",
// u[46279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -5,7 +5,7 @@
//              "LastObjectSize": "361",
//              "ModTime": "55",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:3",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Value": {
//              "T": {
//     @@ -14,7 +14,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "23c79341f8583900b8e4138982dda98181fc2d8c",
//     +            "Hash": "a4a24b877058de8352091982ecaec27ac9000646",
//                  "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:5"
//              }
//          }
// d[a7e973c6aeab1f1cb747d0903cb1e4d369226376:54](-507)
// d[a7e973c6aeab1f1cb747d0903cb1e4d369226376:55](-404)
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
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
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// c[a7e973c6aeab1f1cb747d0903cb1e4d369226376:56](380)={
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:56",
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
//     +            "Hash": "fb1c80f4f55ce8f2e5574fac55204983b23eefb1",
//     +            "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:56"
//              }
//          }
//      }
// u[46279d1e03ecb38b84822f9caf584bd16913470d:7](0)=
//     @@ -12,7 +12,7 @@
//              "LastObjectSize": "384",
//              "ModTime": "55",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:6",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/tests/vm/crossrealm_b",
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]

// Output:
// hello A cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// hello B cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .
// hello C cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .
// hello D cur=gno.land/r/tests/vm/crossrealm_b prev=gno.land/r/tests/vm/crossrealm
// .
