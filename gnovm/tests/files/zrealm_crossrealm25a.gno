// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type Struct struct {
	A int
}

var ourS *Struct

func init() {
	println("s = &Struct{A: 100}")
	s := &Struct{A: 100}
	println("crossrealm_b.SetObject(cross, &s.A)")
	crossrealm_b.SetObject(cross, &s.A)
	println("ourS = s")
	ourS = s
}

func main(cur realm) {
	println("crossrealm_b.IncrementObject(cross)")
	println(crossrealm_b.IncrementObject(cross))
	println(ourS.A)
}

// Output:
// s = &Struct{A: 100}
// crossrealm_b.SetObject(cross, &s.A)
// ourS = s
// crossrealm_b.IncrementObject(cross)
// &(101 int)
// 101

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[46279d1e03ecb38b84822f9caf584bd16913470d:19](0)=
//     @@ -1,7 +1,7 @@
//      {
//          "Fields": [
//              {
//     -            "N": "ZAAAAAAAAAA=",
//     +            "N": "ZQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -12,7 +12,7 @@
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:19",
//              "IsEscaped": true,
//              "LastObjectSize": "181",
//     -        "ModTime": "7",
//     +        "ModTime": "19",
//              "RefCount": "2"
//          }
//      }
// finalizerealm["gno.land/r/crossrealm"]
