// PKGPATH: gno.land/r/zrealm_fvshare_keep
// R2a: a writer already saved is not copied when stored again, so dropping
// one holder leaves the other's object in place.
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
// c[08853ff7484e070ea40ac97921f615b7059ad2ff:56](506)={
//     "Crossing": false,
//     "FileName": "fvshare.gno",
//     "IsClosure": false,
//     "IsMethod": true,
//     "Name": "Bump",
//     "NativeName": "",
//     "NativePkg": "",
//     "ObjectInfo": {
//         "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:56",
//         "LastObjectSize": "506",
//         "ModTime": "0",
//         "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:55",
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
// c[08853ff7484e070ea40ac97921f615b7059ad2ff:55](872)={
//     "Func": {
//         "Crossing": false,
//         "FileName": "fvshare.gno",
//         "IsClosure": false,
//         "IsMethod": true,
//         "Name": "Bump",
//         "NativeName": "",
//         "NativePkg": "",
//         "ObjectInfo": {
//             "Hash": "d3d52540c732ffd73e09f7bc5322bb67477ebe82",
//             "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:56",
//             "LastObjectSize": "506",
//             "ModTime": "0",
//             "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:55",
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
//         "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:55",
//         "LastObjectSize": "872",
//         "ModTime": "0",
//         "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:54",
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
// c[08853ff7484e070ea40ac97921f615b7059ad2ff:54](316)={
//     "ObjectInfo": {
//         "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:54",
//         "LastObjectSize": "316",
//         "ModTime": "0",
//         "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:53",
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
//             "Hash": "72bc7d697c043942d5e2a1bc9fd6489d941bddb5",
//             "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:55"
//         }
//     }
// }
// c[08853ff7484e070ea40ac97921f615b7059ad2ff:53](471)={
//     "Captures": [
//         {
//             "T": {
//                 "@type": "/gno.heapItemType"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "1d23836ad9ad756e73b9cfd88ffe3997768e9064",
//                 "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:54"
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
//         "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:53",
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
//                     "Line": "268"
//                 },
//                 "Num": "0",
//                 "Pos": {
//                     "Column": "9",
//                     "Line": "268"
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
//     +        "ModTime": "52",
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
//     +            "Hash": "8bcb887d4c8fa4691ff570392d58ff3c197a1715",
//     +            "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:53"
//              }
//          }
//      }
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:7](5)=
//     @@ -4,8 +4,8 @@
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:7",
//              "IsEscaped": true,
//              "LastObjectSize": "366",
//     -        "ModTime": "0",
//     -        "RefCount": "32"
//     +        "ModTime": "56",
//     +        "RefCount": "33"
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
//     +        "ModTime": "56",
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
//     +        "ModTime": "56",
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
//              "LastObjectSize": "9990",
//     -        "ModTime": "0",
//     +        "ModTime": "56",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "9722cf44a4198b4d35b798efcac647dc6e520b4a",
//     +                "Hash": "a7646356c29d85c79e96e9eb3590dd3b6e3df735",
//                      "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:3"
//                  }
//              },
//     @@ -1226,7 +1226,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "a42d0970322113189e09f2dd8467a37f7fab3cbb",
//     +                "Hash": "ceacb9b5d1fecc91902a591f9754c1b1e689574b",
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
//     +        "ModTime": "56",
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
//     +            "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:55"
//              }
//          }
//      }
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:55](6)=
//     @@ -63,10 +63,10 @@
//          "MethodPkg": "",
//          "ObjectInfo": {
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:55",
//     +        "IsEscaped": true,
//              "LastObjectSize": "872",
//     -        "ModTime": "0",
//     -        "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:54",
//     -        "RefCount": "1"
//     +        "ModTime": "56",
//     +        "RefCount": "2"
//          },
//          "Receiver": {
//              "T": {
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:54](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:54",
//              "LastObjectSize": "316",
//     -        "ModTime": "0",
//     +        "ModTime": "56",
//              "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:53",
//              "RefCount": "1"
//          },
//     @@ -14,7 +14,6 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "72bc7d697c043942d5e2a1bc9fd6489d941bddb5",
//                  "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:55"
//              }
//          }
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:2](0)=
//     @@ -1226,7 +1226,6 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "ceacb9b5d1fecc91902a591f9754c1b1e689574b",
//                      "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:48"
//                  }
//              },
//     @@ -1236,7 +1235,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "8074cf3ce7bf70558a8e20aa8c01c57d0419a3ee",
//     +                "Hash": "67dca7269c213320930cc6a472bd292e394ff1f3",
//                      "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:49"
//                  }
//              },
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:53](5)=
//     @@ -6,7 +6,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1d23836ad9ad756e73b9cfd88ffe3997768e9064",
//     +                "Hash": "f2c092642b36354d3593f891938a14183353aafe",
//                      "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:54"
//                  }
//              }
//     @@ -21,7 +21,7 @@
//          "ObjectInfo": {
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:53",
//              "LastObjectSize": "471",
//     -        "ModTime": "0",
//     +        "ModTime": "56",
//              "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:48",
//              "RefCount": "1"
//          },
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:48](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08853ff7484e070ea40ac97921f615b7059ad2ff:48",
//              "LastObjectSize": "349",
//     -        "ModTime": "52",
//     +        "ModTime": "56",
//              "OwnerID": "08853ff7484e070ea40ac97921f615b7059ad2ff:2",
//              "RefCount": "1"
//          },
//     @@ -26,7 +26,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "8bcb887d4c8fa4691ff570392d58ff3c197a1715",
//     +            "Hash": "1a98131aa0b3a67c8c30a3b981450e3c249f3999",
//                  "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:53"
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
//     -            "Hash": "1a98131aa0b3a67c8c30a3b981450e3c249f3999",
//     -            "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:53"
//              }
//          }
//      }
// u[08853ff7484e070ea40ac97921f615b7059ad2ff:55](0)=
//     @@ -66,7 +66,7 @@
//              "IsEscaped": true,
//              "LastObjectSize": "878",
//              "ModTime": "56",
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
//     +                "Hash": "32d6f4beeeee21387c529e3c7c1aa7ae45b54200",
//                      "ObjectID": "08853ff7484e070ea40ac97921f615b7059ad2ff:48"
//                  }
//              },
// d[08853ff7484e070ea40ac97921f615b7059ad2ff:53](-476)
// d[08853ff7484e070ea40ac97921f615b7059ad2ff:54](-321)
// finalizerealm["gno.land/r/zrealm_fvshare_keep"]
