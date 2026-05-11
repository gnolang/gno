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
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4](0)=
//     @@ -23,6 +23,7 @@
//                  }
//              },
//              {
//     +            "N": "BAAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3](0)=
//     @@ -19,7 +19,7 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "b4f1b00926cf2bf163e68f49008c87b915755a13",
//     +                "Hash": "c6307e3a8cb600571915e684b01701e4be55213e",
//                      "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:4"
//                  },
//                  "Length": "0",
// u[09b6b63cc3f789fda754d95769bbf99bf64c5cb4:2](0)=
//     @@ -34,7 +34,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "9db05fac726a131bb3ed18dc13dfdbe813f3f5f9",
//     +                "Hash": "4ce456d5c90e31a9b2088ab592f66181ec2fe201",
//                      "ObjectID": "09b6b63cc3f789fda754d95769bbf99bf64c5cb4:3"
//                  }
//              },
