// PKGPATH: gno.land/r/zrealm_fvshare_keep
// A writer already saved and stored again keeps one object, so dropping one
// holder leaves the other's object in place.
package zrealm_fvshare_keep

import "gno.land/r/tests/vm/fvshare"

func main(cur realm) {
	fvshare.MakeKeep(cross(cur))
	fvshare.UseKeep(cross(cur))
	fvshare.ClearKeep(cross(cur))
	println("done")
}

// Output:
// done

// Realm:
// finalizerealm["gno.land/r/tests/vm/fvshare"]
// c[08853ff7484e070ea40ac97921f615b7059ad2ff:77](506)={
//     "Crossing": false,
//     "FileName": "fvshare.gno",
//     "IsClosure": false,
//     "IsMethod": true,
//     "Name": "Bump",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:77",
//         "LastObjectSize": "506",
//         "ModTime": "0",
//         "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:76",
//         "RefCount": "1"
//     },
//     "Parent": {
//         "@type": "/gno.RefValue",
//         "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:7"
//     },
//     "PkgPath": "gno.land/r/tests/vm/fvshare",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "fvshare.gno",
//             "PkgPath": "gno.land/r/tests/vm/fvshare",
//             "Span": {
//                 "End": {
//                     "Column": "33",
//                     "Line": "9"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "1",
//                     "Line": "9"
//                 }
//             }
//         }
//     },
//     "Type": {
//         "@type": "/gno.FuncType",
//         "Params": [
//             {
//                 "Embedded": false,
//                 "Name": "v",
//                 "PkgPath": "",
//                 "Tag": "",
//                 "Type": {
//                     "@type": "/gno.PointerType",
//                     "Elt": {
//                         "@type": "/gno.RefType",
//                         "ID": "gno.land/r/tests/vm/fvshare.Vault"
//                     }
//                 }
//             }
//         ],
//         "Results": []
//     }
// }
// c[08853ff7484e070ea40ac97921f615b7059ad2ff:76](872)={
//     "Func": {
//         "Crossing": false,
//         "FileName": "fvshare.gno",
//         "IsClosure": false,
//         "IsMethod": true,
//         "Name": "Bump",
//         "NativeName": "",
//         "NativePkg": "",
//         "ObjectInfo": {
//             "Hash": "6fd365a17c4cf04074e88c875117facb98af35c0",
//             "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:77",
//             "LastObjectSize": "506",
//             "ModTime": "0",
//             "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:76",
//             "RefCount": "1"
//         },
//         "Parent": {
//             "@type": "/gno.RefValue",
//             "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:7"
//         },
//         "PkgPath": "gno.land/r/tests/vm/fvshare",
//         "Source": {
//             "@type": "/gno.RefNode",
//             "BlockNode": null,
//             "Location": {
//                 "File": "fvshare.gno",
//                 "PkgPath": "gno.land/r/tests/vm/fvshare",
//                 "Span": {
//                     "End": {
//                         "Column": "33",
//                         "Line": "9"
//                     },
//                     "Num": "0",
//                     "Pos": {
//                         "Column": "1",
//                         "Line": "9"
//                     }
//                 }
//             }
//         },
//         "Type": {
//             "@type": "/gno.FuncType",
//             "Params": [
//                 {
//                     "Embedded": false,
//                     "Name": "v",
//                     "PkgPath": "",
//                     "Tag": "",
//                     "Type": {
//                         "@type": "/gno.PointerType",
//                         "Elt": {
//                             "@type": "/gno.RefType",
//                             "ID": "gno.land/r/tests/vm/fvshare.Vault"
//                         }
//                     }
//                 }
//             ],
//             "Results": []
//         }
//     },
//     "Method": "",
//     "MethodPkg": "",
//     "ObjectInfo": {
//         "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:76",
//         "LastObjectSize": "872",
//         "ModTime": "0",
//         "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:75",
//         "RefCount": "1"
//     },
//     "Receiver": {
//         "T": {
//             "@type": "/gno.PointerType",
//             "Elt": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/tests/vm/fvshare.Vault"
//             }
//         },
//         "V": {
//             "@type": "/gno.PointerValue",
//             "Base": {
//                 "@type": "/gno.RefValue",
//                 "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:4"
//             },
//             "Index": "0",
//             "TV": null
//         }
//     }
// }
// c[08853ff7484e070ea40ac97921f615b7059ad2ff:75](316)={
//     "ObjectInfo": {
//         "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:75",
//         "LastObjectSize": "316",
//         "ModTime": "0",
//         "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:74",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.FuncType",
//             "Params": [],
//             "Results": []
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "d57a0df05335d0b5855f74666d06f801c499a7ee",
//             "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:76"
//         }
//     }
// }
// c[08853ff7484e070ea40ac97921f615b7059ad2ff:74](471)={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "369ead352e894e2500679d0f2477eac0e532fecf",
//                 "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:75"
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
//         "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:74",
//         "LastObjectSize": "471",
//         "ModTime": "0",
//         "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:48",
//         "RefCount": "1"
//     },
//     "Parent": null,
//     "PkgPath": "gno.land/r/tests/vm/fvshare",
//     "Source": {
//         "@type": "/gno.RefNode",
//         "BlockNode": null,
//         "Location": {
//             "File": "fvshare.gno",
//             "PkgPath": "gno.land/r/tests/vm/fvshare",
//             "Span": {
//                 "End": {
//                     "Column": "35",
//                     "Line": "270"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "9",
//                     "Line": "270"
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
//                     "@type": "/gno.FuncType",
//                     "Params": [],
//                     "Results": []
//                 }
//             }
//         ]
//     }
// }
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:48](112)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:48",
//              "LastObjectSize": "237",
//     -        "ModTime": "0",
//     +        "ModTime": "73",
//              "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:2",
//              "RefCount": "1"
//          },
//     @@ -23,6 +23,11 @@
//                          }
//                      }
//                  ]
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "2dff5bfee1c0655ed28c4707c5b4fa9d7718aad3",
//     +            "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:74"
//              }
//          }
//      }
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:7](5)=
//     @@ -4,8 +4,8 @@
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:7",
//              "IsEscaped": true,
//              "LastObjectSize": "366",
//     -        "ModTime": "0",
//     -        "RefCount": "52"
//     +        "ModTime": "77",
//     +        "RefCount": "53"
//          },
//          "Parent": {
//              "@type": "/gno.RefValue",
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:4](7)=
//     @@ -1,10 +1,10 @@
//      {
//          "ObjectInfo": {
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:4",
//     +        "IsEscaped": true,
//              "LastObjectSize": "349",
//     -        "ModTime": "0",
//     -        "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:3",
//     -        "RefCount": "1"
//     +        "ModTime": "77",
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:3",
//              "LastObjectSize": "394",
//     -        "ModTime": "0",
//     +        "ModTime": "77",
//              "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,6 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "9a71fc6d124c7e24d6741a49f05c3a253b24054a",
//                      "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:4"
//                  },
//                  "Index": "0",
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:2](6)=
//     @@ -4,7 +4,7 @@
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:2",
//              "IsEscaped": true,
//              "LastObjectSize": "16008",
//     -        "ModTime": "0",
//     +        "ModTime": "77",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "9722cf44a4198b4d35b798efcac647dc6e520b4a",
//     +                "Hash": "8696574ddb55cfb39a54fa346469ad8f869f27fc",
//                      "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:3"
//                  }
//              },
//     @@ -1226,7 +1226,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "a42d0970322113189e09f2dd8467a37f7fab3cbb",
//     +                "Hash": "d81624418c7d0d50417e5f123896f0e621c77eb7",
//                      "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:48"
//                  }
//              },
// finalizerealm["gno.land/r/tests/vm/fvshare"]
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:49](112)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:49",
//              "LastObjectSize": "208",
//     -        "ModTime": "0",
//     +        "ModTime": "77",
//              "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:2",
//              "RefCount": "1"
//          },
//     @@ -11,6 +11,10 @@
//                  "@type": "/gno.FuncType",
//                  "Params": null,
//                  "Results": null
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:76"
//              }
//          }
//      }
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:76](6)=
//     @@ -63,10 +63,10 @@
//          "MethodPkg": "",
//          "ObjectInfo": {
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:76",
//     +        "IsEscaped": true,
//              "LastObjectSize": "872",
//     -        "ModTime": "0",
//     -        "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:75",
//     -        "RefCount": "1"
//     +        "ModTime": "77",
//     +        "RefCount": "2"
//          },
//          "Receiver": {
//              "T": {
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:75](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:75",
//              "LastObjectSize": "316",
//     -        "ModTime": "0",
//     +        "ModTime": "77",
//              "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:74",
//              "RefCount": "1"
//          },
//     @@ -14,7 +14,6 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "d57a0df05335d0b5855f74666d06f801c499a7ee",
//                  "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:76"
//              }
//          }
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:2](0)=
//     @@ -1226,7 +1226,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "d81624418c7d0d50417e5f123896f0e621c77eb7",
//                      "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:48"
//                  }
//              },
//     @@ -1236,7 +1235,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "8074cf3ce7bf70558a8e20aa8c01c57d0419a3ee",
//     +                "Hash": "1222686475db843dc062cd725e4d0db3a14e5884",
//                      "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:49"
//                  }
//              },
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:74](5)=
//     @@ -6,7 +6,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "369ead352e894e2500679d0f2477eac0e532fecf",
//     +                "Hash": "22914d296b27af916196641da8ee4e53a2b1614d",
//                      "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:75"
//                  }
//              }
//     @@ -21,7 +21,7 @@
//          "ObjectInfo": {
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:74",
//              "LastObjectSize": "471",
//     -        "ModTime": "0",
//     +        "ModTime": "77",
//              "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:48",
//              "RefCount": "1"
//          },
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:48](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:48",
//              "LastObjectSize": "349",
//     -        "ModTime": "73",
//     +        "ModTime": "77",
//              "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:2",
//              "RefCount": "1"
//          },
//     @@ -26,7 +26,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "2dff5bfee1c0655ed28c4707c5b4fa9d7718aad3",
//     +            "Hash": "f010d67401cb8fddb06af2436a6718cbd58fea23",
//                  "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:74"
//              }
//          }
// finalizerealm["gno.land/r/tests/vm/fvshare"]
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:48](-107)=
//     @@ -23,11 +23,6 @@
//                          }
//                      }
//                  ]
//     -        },
//     -        "V": {
//     -            "@type": "/gno.RefValue",
//     -            "Hash": "f010d67401cb8fddb06af2436a6718cbd58fea23",
//     -            "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:74"
//              }
//          }
//      }
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:76](0)=
//     @@ -66,7 +66,7 @@
//              "IsEscaped": true,
//              "LastObjectSize": "878",
//              "ModTime": "77",
//     -        "RefCount": "2"
//     +        "RefCount": "1"
//          },
//          "Receiver": {
//              "T": {
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:2](0)=
//     @@ -1226,6 +1226,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     +                "Hash": "4a5f25b5b73e2808f1764cf8055a3c5759b13b7a",
//                      "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:48"
//                  }
//              },
// d[08853ff7484e070ea40ac97921f615b7059ad2ff:74](-476)
// d[08853ff7484e070ea40ac97921f615b7059ad2ff:75](-321)
// finalizerealm["gno.land/r/zrealm_fvshare_keep"]
