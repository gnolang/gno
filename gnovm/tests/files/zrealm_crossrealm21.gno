// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm"
	"gno.land/r/tests/vm/crossrealm_b"
)

func main(cur realm) {
	f := crossrealm_b.Fooer
	crossrealm.SetFooer(cross, f)
	crossrealm.CallFooerFooCross(cross)
	f.SetS("B")
	crossrealm.CallFooerFooCross(cross)
	print(".")
}

<<<<<<< HEAD
//     +        "ModTime": "53",
//              "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
=======
//     +        "ModTime": "43",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:3",
>>>>>>> main/master
//     -        "RefCount": "1"
//     +        "RefCount": "2"
//          },
//          "Value": {
//              "T": {
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[46279d1e03ecb38b84822f9caf584bd16913470d:5](5)=
//     @@ -7,14 +7,14 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "A"
//     +                "value": "B"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:5",
//              "LastObjectSize": "239",
//     -        "ModTime": "0",
//     +        "ModTime": "15",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:4",
//              "RefCount": "1"
//          }
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// finalizerealm["gno.land/r/tests/vm/crossrealm"]
// finalizerealm["gno.land/r/crossrealm"]

// Error:
// zrealm_crossrealm21.gno:18:1: expected declaration, found '<<'

// TypeCheckError:
// gno.land/r/crossrealm/zrealm_crossrealm21.gno:18:1: expected declaration, found '<<'
