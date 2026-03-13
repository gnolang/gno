// PKGPATH: gno.land/r/xx
package xx

var i int8
var is []int8
var s string
var ss []string

func main() {
	i = 127
	is = []int8{0, 127}
	s = "a"
	ss = []string{"a", "aa"}

}

// Storage:
// gno.land/r/xx: 871

// Realm:
// finalizerealm["gno.land/r/xx"]
// c[aea84df38908f9569d0f552575606e6e6e7e22dd:9](253)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "fwAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:9",
//         "LastObjectSize": "253",
//         "ModTime": "0",
//         "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:4",
//         "RefCount": "1"
//     }
// }
// c[aea84df38908f9569d0f552575606e6e6e7e22dd:10](303)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "a"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "aa"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:10",
//         "LastObjectSize": "303",
//         "ModTime": "0",
//         "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:6",
//         "RefCount": "1"
//     }
// }
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:3](5)=
//     @@ -2,11 +2,12 @@
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:3",
//              "LastObjectSize": "217",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "RefCount": "1"
//          },
//          "Value": {
//     +        "N": "fwAAAAAAAAA=",
//              "T": {
//                  "@type": "/gno.PrimitiveType",
//                  "value": "64"
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:4](137)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:4",
//              "LastObjectSize": "237",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "RefCount": "1"
//          },
//     @@ -14,6 +14,17 @@
//                      "value": "64"
//                  },
//                  "Vrd": false
//     +        },
//     +        "V": {
//     +            "@type": "/gno.SliceValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "e84e8bceb6edffed359aabd64d88ad4cdaae23c8",
//     +                "ObjectID": "aea84df38908f9569d0f552575606e6e6e7e22dd:9"
//     +            },
//     +            "Length": "2",
//     +            "Maxcap": "2",
//     +            "Offset": "0"
//              }
//          }
//      }
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:5](30)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:5",
//              "LastObjectSize": "216",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "RefCount": "1"
//          },
//     @@ -10,6 +10,10 @@
//              "T": {
//                  "@type": "/gno.PrimitiveType",
//                  "value": "16"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.StringValue",
//     +            "value": "a"
//              }
//          }
//      }
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:6](138)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:6",
//              "LastObjectSize": "236",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "RefCount": "1"
//          },
//     @@ -14,6 +14,17 @@
//                      "value": "16"
//                  },
//                  "Vrd": false
//     +        },
//     +        "V": {
//     +            "@type": "/gno.SliceValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "af828a01b9823a07083d08cb1c69bbe1f87c0edb",
//     +                "ObjectID": "aea84df38908f9569d0f552575606e6e6e7e22dd:10"
//     +            },
//     +            "Length": "2",
//     +            "Maxcap": "2",
//     +            "Offset": "0"
//              }
//          }
//      }
// u[aea84df38908f9569d0f552575606e6e6e7e22dd:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "aea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "IsEscaped": true,
//              "LastObjectSize": "910",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5dcd10eb5c8e46affb96b282cd33f99e395ca03f",
//     +                "Hash": "50c30f49a200d55ca944f7b558ad73afad257bec",
//                      "ObjectID": "aea84df38908f9569d0f552575606e6e6e7e22dd:3"
//                  }
//              },
//     @@ -44,7 +44,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "778fc5bbb8f67c7d420cb4985661bc5619a4c847",
//     +                "Hash": "b13dbd1115ddd93332215c7385f0b3e9e78106fa",
//                      "ObjectID": "aea84df38908f9569d0f552575606e6e6e7e22dd:4"
//                  }
//              },
//     @@ -54,7 +54,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "4fc3275a1a3adaff472b6a574029f4b55b7c81d3",
//     +                "Hash": "5c355c4f673df9e920bd5a343ef511ea5470cc12",
//                      "ObjectID": "aea84df38908f9569d0f552575606e6e6e7e22dd:5"
//                  }
//              },
//     @@ -64,7 +64,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "36036c7b28ac2bfb8cec7f49392e3a6c3621a31f",
//     +                "Hash": "3e9876855fa3a5fd1e2138b653cde45481b549cf",
//                      "ObjectID": "aea84df38908f9569d0f552575606e6e6e7e22dd:6"
//                  }
//              },
