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
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:21](215)={
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
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:21",
//         "LastObjectSize": "215",
//         "ModTime": "0",
//         "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//         "RefCount": "1"
//     }
// }
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:22](215)={
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
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:22",
//         "LastObjectSize": "215",
//         "ModTime": "0",
//         "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//         "RefCount": "1"
//     }
// }
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:23](215)={
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
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:23",
//         "LastObjectSize": "215",
//         "ModTime": "0",
//         "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//         "RefCount": "1"
//     }
// }
// c[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:24](215)={
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
//         "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:24",
//         "LastObjectSize": "215",
//         "ModTime": "0",
//         "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//         "RefCount": "1"
//     }
// }
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4](1)=
//     @@ -8,8 +8,8 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "b64e6d66625500a1a73ad3aa83bf64a19e5ef79c",
//     -                "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:18"
//     +                "Hash": "a50095053b0560d6000a853698d2a70398b43ff0",
//     +                "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:21"
//                  }
//              },
//              {
//     @@ -19,8 +19,8 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "044a7cc5cdc8d198ee0de4c97b17542ba2bcc45d",
//     -                "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:19"
//     +                "Hash": "4bd56d347d1de8616331e0c7e9d24f2e660b7506",
//     +                "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:22"
//                  }
//              },
//              {
//     @@ -30,8 +30,8 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "c2c402289358af04b4bcc1263a34b23304591963",
//     -                "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:20"
//     +                "Hash": "0fd141d953069d60bf97cb84881f33976f022e11",
//     +                "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:23"
//                  }
//              },
//              {
//     @@ -41,8 +41,8 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "6f563d96f793175f993169399dc7a9b7ff7f5682",
//     -                "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:8"
//     +                "Hash": "34ef9d7fdb562e442c9fef95bc3c06f7613b119b",
//     +                "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:24"
//                  }
//              },
//              {
//     @@ -115,7 +115,7 @@
//          "ObjectInfo": {
//              "ID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4",
//              "LastObjectSize": "1738",
//     -        "ModTime": "17",
//     +        "ModTime": "20",
//              "OwnerID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3",
//              "RefCount": "1"
//          }
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:18](-215)
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:19](-215)
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:20](-215)
// d[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:8](-214)
