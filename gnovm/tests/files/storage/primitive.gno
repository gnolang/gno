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
// c[0ea84df38908f9569d0f552575606e6e6e7e22dd:9](253)={
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
//         "ID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:9",
//         "LastObjectSize": "253",
//         "ModTime": "0",
//         "OwnerID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:4",
//         "RefCount": "1"
//     }
// }
// c[0ea84df38908f9569d0f552575606e6e6e7e22dd:10](303)={
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
//         "ID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:10",
//         "LastObjectSize": "303",
//         "ModTime": "0",
//         "OwnerID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:6",
//         "RefCount": "1"
//     }
// }
// u[0ea84df38908f9569d0f552575606e6e6e7e22dd:3](5)=
//     @@ -2,11 +2,12 @@
//          "ObjectInfo": {
//              "ID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:3",
//              "LastObjectSize": "217",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:2",
//              "RefCount": "1"
//          },
//          "Value": {
//     +        "N": "fwAAAAAAAAA=",
//              "T": {
//                  "@type": "/gno.PrimitiveType",
//                  "value": "64"
// u[0ea84df38908f9569d0f552575606e6e6e7e22dd:4](137)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:4",
//              "LastObjectSize": "237",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:2",
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
//     +                "Hash": "d1bfe56cd0fa827877a05e18bb530523a5290dc5",
//     +                "ObjectID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:9"
//     +            },
//     +            "Length": "2",
//     +            "Maxcap": "2",
//     +            "Offset": "0"
//              }
//          }
//      }
// u[0ea84df38908f9569d0f552575606e6e6e7e22dd:5](30)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:5",
//              "LastObjectSize": "216",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:2",
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
// u[0ea84df38908f9569d0f552575606e6e6e7e22dd:6](138)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:6",
//              "LastObjectSize": "236",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:2",
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
//     +                "Hash": "0b321de960853f22df9d4635edffd1990257bbfc",
//     +                "ObjectID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:10"
//     +            },
//     +            "Length": "2",
//     +            "Maxcap": "2",
//     +            "Offset": "0"
//              }
//          }
//      }
// u[0ea84df38908f9569d0f552575606e6e6e7e22dd:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:2",
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
//     -                "Hash": "48fbf787cdabaa8a9f3178db52d7f4179244f8ce",
//     +                "Hash": "3e1611b9202c0ebcf52e77afa4429246ec4ed9df",
//                      "ObjectID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:3"
//                  }
//              },
//     @@ -44,7 +44,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "c331a30779947abf214738bd50f7b5d5b4bf9a7c",
//     +                "Hash": "88cf2834f0746a10f61d650e38d4b4d2d8850ce7",
//                      "ObjectID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:4"
//                  }
//              },
//     @@ -54,7 +54,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e263e5aa92eb546793084471e03c529f1f3596d8",
//     +                "Hash": "a1593b748576f13e1a3b6a32b64c78b7098788a1",
//                      "ObjectID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:5"
//                  }
//              },
//     @@ -64,7 +64,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "923bd5f136668a45c32e2f313b98f335549325e7",
//     +                "Hash": "0d3e101a6f9f41e8ddab7a17ea54c2c2decd4df5",
//                      "ObjectID": "0ea84df38908f9569d0f552575606e6e6e7e22dd:6"
//                  }
//              },
