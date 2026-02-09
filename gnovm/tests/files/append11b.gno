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
