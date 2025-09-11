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

<<<<<<< HEAD
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:54](509)={
=======
// c[a7e973c6aeab1f1cb747d0903cb1e4d369226376:44](507)={
>>>>>>> main/master
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
<<<<<<< HEAD
//                 "Hash": "6add0a2ec2762807781a0120c0dca200016f3ffc",
//                 "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:55"
=======
//                 "Hash": "49300d9e118dce954496bef87e1ab6b09625b24d",
//                 "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:45"
>>>>>>> main/master
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
<<<<<<< HEAD
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:54",
//         "LastObjectSize": "509",
=======
//         "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:44",
//         "LastObjectSize": "507",
>>>>>>> main/master
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
<<<<<<< HEAD
//     +        "ModTime": "53",
//              "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
=======
//     +        "ModTime": "43",
//              "OwnerID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:2",
>>>>>>> main/master
//              "RefCount": "1"
//          },
//     @@ -10,6 +10,11 @@
//              "T": {
//                  "@type": "/gno.RefType",
//                  "ID": "gno.land/r/tests/vm/crossrealm.FooerGetter"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
<<<<<<< HEAD
//     +            "Hash": "35dcfd645127b32187a91fafc2061ababf14059b",
//     +            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:54"
=======
//     +            "Hash": "22e2f7ebcea372457a1a731625f214583f9a14a5",
//     +            "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:44"
>>>>>>> main/master
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
<<<<<<< HEAD
//     +        "ModTime": "55",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
=======
//     +        "ModTime": "45",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:3",
>>>>>>> main/master
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
<<<<<<< HEAD
// finalizerealm["gno.land/r/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:5](5)=
=======
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[46279d1e03ecb38b84822f9caf584bd16913470d:5](5)=
>>>>>>> main/master
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
<<<<<<< HEAD
// finalizerealm["gno.land/r/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:5](0)=
=======
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[46279d1e03ecb38b84822f9caf584bd16913470d:5](0)=
>>>>>>> main/master
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
<<<<<<< HEAD
//              "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:14",
//              "LastObjectSize": "367",
//     -        "ModTime": "53",
//     +        "ModTime": "55",
//              "OwnerID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:2",
=======
//              "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:14",
//              "LastObjectSize": "365",
//     -        "ModTime": "43",
//     +        "ModTime": "45",
//              "OwnerID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:2",
>>>>>>> main/master
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
<<<<<<< HEAD
//     -            "Hash": "35dcfd645127b32187a91fafc2061ababf14059b",
//     -            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:54"
=======
//     -            "Hash": "22e2f7ebcea372457a1a731625f214583f9a14a5",
//     -            "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:44"
>>>>>>> main/master
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
<<<<<<< HEAD
//     +        "ModTime": "55",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:6",
=======
//     +        "ModTime": "45",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:6",
>>>>>>> main/master
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Parent": null,
//          "PkgPath": "gno.land/r/tests/vm/crossrealm_b",
// u[46279d1e03ecb38b84822f9caf584bd16913470d:4](0)=
//     @@ -5,7 +5,7 @@
<<<<<<< HEAD
//              "LastObjectSize": "363",
//              "ModTime": "55",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
=======
//              "LastObjectSize": "361",
//              "ModTime": "45",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:3",
>>>>>>> main/master
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
<<<<<<< HEAD
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:54](-509)
// d[1712ac7adcfdc8e58a67e5615e20fb312394c4df:55](-406)
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:5](0)=
=======
// d[a7e973c6aeab1f1cb747d0903cb1e4d369226376:44](-507)
// d[a7e973c6aeab1f1cb747d0903cb1e4d369226376:45](-404)
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[46279d1e03ecb38b84822f9caf584bd16913470d:5](0)=
>>>>>>> main/master
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "C"
//     +                "value": "D"
//                  }
//              }
//          ],
<<<<<<< HEAD
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
// c[1712ac7adcfdc8e58a67e5615e20fb312394c4df:56](386)={
=======
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// c[a7e973c6aeab1f1cb747d0903cb1e4d369226376:46](380)={
>>>>>>> main/master
//     "Crossing": false,
//     "FileName": "",
//     "IsClosure": false,
//     "IsMethod": false,
//     "Name": "",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
<<<<<<< HEAD
//         "ID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:56",
//         "LastObjectSize": "386",
=======
//         "ID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:46",
//         "LastObjectSize": "380",
>>>>>>> main/master
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
<<<<<<< HEAD
//     -            "ObjectID": "0edc46caf30c00efd87b6c272673239eafbd051e:7"
//     +            "Hash": "0bb5b4426473c9939bf53c492d8f92f4bd48dfdb",
//     +            "ObjectID": "1712ac7adcfdc8e58a67e5615e20fb312394c4df:56"
=======
//     -            "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:7"
//     +            "Hash": "cf4b669212d7a060336ace4ecf29fb8c45b0a195",
//     +            "ObjectID": "a7e973c6aeab1f1cb747d0903cb1e4d369226376:46"
>>>>>>> main/master
//              }
//          }
//      }
// u[46279d1e03ecb38b84822f9caf584bd16913470d:7](0)=
//     @@ -12,7 +12,7 @@
<<<<<<< HEAD
//              "LastObjectSize": "390",
//              "ModTime": "55",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:6",
=======
//              "LastObjectSize": "384",
//              "ModTime": "45",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:6",
>>>>>>> main/master
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Parent": null,
<<<<<<< HEAD
//          "PkgPath": "gno.land/r/demo/tests/crossrealm_b",
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm_b"]
// finalizerealm["gno.land/r/demo/tests/crossrealm"]
=======
//          "PkgPath": "gno.land/r/tests/vm/crossrealm_b",
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
>>>>>>> main/master
// finalizerealm["gno.land/r/crossrealm"]

// Error:
// zrealm_crossrealm22.gno:28:1: expected declaration, found '<<'

// TypeCheckError:
// gno.land/r/crossrealm/zrealm_crossrealm22.gno:28:1: expected declaration, found '<<'
