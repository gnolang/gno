// PKGPATH: gno.land/r/main
package main

import "fmt"

type S struct {
	A int
}

var bs []S = make([]S, 0, 10)

func init() {
	_ = append(bs, S{1}, S{2}, S{3})
}

func main() {
	fmt.Println(bs[0:3])
	_ = append(bs, S{1}, S{2}, S{3}, S{4})
	fmt.Println(bs[0:4])
}

// Output:
// [{1} {2} {3}]
// [{1} {2} {3} {4}]

// Realm:
// finalizerealm["gno.land/r/main"]
// c[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:21](215)={
//     "Fields": [
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:21",
//         "LastObjectSize": "215",
//         "ModTime": "0",
//         "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//         "RefCount": "1"
//     }
// }
// c[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:22](215)={
//     "Fields": [
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:22",
//         "LastObjectSize": "215",
//         "ModTime": "0",
//         "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//         "RefCount": "1"
//     }
// }
// c[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:23](215)={
//     "Fields": [
//         {
//             "N": "AwAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:23",
//         "LastObjectSize": "215",
//         "ModTime": "0",
//         "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//         "RefCount": "1"
//     }
// }
// c[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:24](215)={
//     "Fields": [
//         {
//             "N": "BAAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:24",
//         "LastObjectSize": "215",
//         "ModTime": "0",
//         "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//         "RefCount": "1"
//     }
// }
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4](1)=
//     @@ -8,8 +8,8 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "3d3f4fca85becfd2f6b8d815f84f4ff8aac359fb",
//     -                "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:18"
//     +                "Hash": "53d69e66700d965f7148cf885d0259ec9ec5c3b7",
//     +                "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:21"
//                  }
//              },
//              {
//     @@ -19,8 +19,8 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "d50b7a33776f6f1c1d757b8ea75a599c131be937",
//     -                "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:19"
//     +                "Hash": "fa78b6ce0e48cfb542702f012c6bc6fce8be6c66",
//     +                "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:22"
//                  }
//              },
//              {
//     @@ -30,8 +30,8 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "70bca675775dd6dcc49de4f0ea7dfd06657aa8b9",
//     -                "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:20"
//     +                "Hash": "8e34dadb02e73e22238ab3d889a518d1e66aa77a",
//     +                "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:23"
//                  }
//              },
//              {
//     @@ -41,8 +41,8 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "9c241ce5ac2a49e7c0268f64c07b42ad36cdc68f",
//     -                "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:8"
//     +                "Hash": "f1d8c67694bafee5f9aa1b2302e3a154ab25a352",
//     +                "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:24"
//                  }
//              },
//              {
//     @@ -115,7 +115,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "LastObjectSize": "1738",
//     -        "ModTime": "17",
//     +        "ModTime": "20",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "RefCount": "1"
//          }
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "LastObjectSize": "381",
//     -        "ModTime": "20",
//     +        "ModTime": "24",
//              "OwnerID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2",
//              "RefCount": "1"
//          },
//     @@ -19,7 +19,7 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "453a5c1681d895af2a48debff58a8680a64ffb1b",
//     +                "Hash": "c33d0909091102d3aeb32c4a55bad0aa2601403c",
//                      "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4"
//                  },
//                  "Length": "0",
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2",
//              "IsEscaped": true,
//              "LastObjectSize": "723",
//     -        "ModTime": "20",
//     +        "ModTime": "24",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "883404846c849c393a0d035a0a4319936507e13a",
//     +                "Hash": "dd0159859ed35b61f2b2d662422e13574a3bb996",
//                      "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3"
//                  }
//              },
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:18](-215)
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:19](-215)
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:20](-215)
// d[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:8](-214)
