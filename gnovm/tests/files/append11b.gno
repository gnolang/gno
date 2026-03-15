// PKGPATH: gno.land/r/main
package main

import "fmt"

var bs []int = make([]int, 0, 10)

func init() {
	_ = append(bs, 1, 2, 3)
}

func main() {
	fmt.Println(bs[0:3])
	_ = append(bs, 1, 2, 3, 4)
	fmt.Println(bs[0:4])
}

// Output:
// [1 2 3]
// [1 2 3 4]

// Realm:
// finalizerealm["gno.land/r/main"]
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4](0)=
//     @@ -23,6 +23,7 @@
//                  }
//              },
//              {
//     +            "N": "BAAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3](0)=
//     @@ -19,7 +19,7 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "ca34ecb7af3bb7088607c1902b26b57e7e89e484",
//     +                "Hash": "253f8d7cad796c300d6cb437a3f6f5cc3b989164",
//                      "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:4"
//                  },
//                  "Length": "0",
// u[f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "026913757278840a39247b2096e8fa9982fcdb08",
//     +                "Hash": "4f77cc6ffe62fefe3b13a46c0958556b02eecdbe",
//                      "ObjectID": "f9b6b63cc3f789fda754d95769bbf99bf64c5cb4:3"
//                  }
//              },
