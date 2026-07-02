// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type Struct struct {
	A int
}

var s *Struct

func init(cur realm) {
	s = &Struct{A: 100}
	crossrealm_b.SetObject(cross(cur), func() { println(&s.A) })
}

func main(cur realm) {
	//s := &Struct{A: 100}
	crossrealm_b.SetObject(cross(cur), 123)
	s.A = 123
	println(s)
}

// Output:
// &(struct{(123 int)} gno.land/r/crossrealm.Struct)

// Preprocessed:
// file{ package crossrealm; import crossrealm_b gno.land/r/tests/vm/crossrealm_b; type Struct (const-type gno.land/r/crossrealm.Struct); var s<!~VPBlock(2,1)> *(typeval{gno.land/r/crossrealm.Struct}); func init.2(cur (const-type .uverse.realm)) { s<~VPBlock(3,1)> = &((const-type gno.land/r/crossrealm.Struct){A<VPField(0,0,A)>: (const (100 int))}); (const (ref(gno.land/r/tests/vm/crossrealm_b) package{})).SetObject((const (cross func(.uverse.realm) .uverse.realm))(cur<VPBlock(1,0)>), func func(){ (const (println func(...interface {})))(&((const (ref(gno.land/r/crossrealm) package{})).s.A)) }) }; func main(cur (const-type .uverse.realm)) { (const (ref(gno.land/r/tests/vm/crossrealm_b) package{})).SetObject((const (cross func(.uverse.realm) .uverse.realm))(cur<VPBlock(1,0)>), (const (123 int))); s<~VPBlock(3,1)>.A = (const (123 int)); (const (println func(...interface {})))(s<~VPBlock(3,1)>) } }

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[06279d1e03ecb38b84822f9caf584bd16913470d:13](-97)=
//     @@ -7,15 +7,10 @@
//              "RefCount": "1"
//          },
//          "Value": {
//     +        "N": "ewAAAAAAAAA=",
//              "T": {
//     -            "@type": "/gno.FuncType",
//     -            "Params": null,
//     -            "Results": null
//     -        },
//     -        "V": {
//     -            "@type": "/gno.RefValue",
//     -            "Hash": "5c3b4df4ca7c54a7afa7d5c4fae642c0bc7e5008",
//     -            "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:7"
//     +            "@type": "/gno.PrimitiveType",
//     +            "value": "32"
//              }
//          }
//      }
// u[06279d1e03ecb38b84822f9caf584bd16913470d:2](0)=
//     @@ -118,7 +118,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "cc740066e57cf339d47789a0a65be2e547386df5",
//     +                "Hash": "a0d0fc341bf851d44366422fc274c3988cb7fe63",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:13"
//                  }
//              },
// d[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:7](-299)
// finalizerealm["gno.land/r/crossrealm"]
// u[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:9](5)=
//     @@ -1,7 +1,7 @@
//      {
//          "Fields": [
//              {
//     -            "N": "ZAAAAAAAAAA=",
//     +            "N": "ewAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -11,7 +11,7 @@
//          "ObjectInfo": {
//              "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:9",
//              "LastObjectSize": "214",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8",
//              "RefCount": "1"
//          }
// u[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8",
//              "LastObjectSize": "344",
//     -        "ModTime": "0",
//     +        "ModTime": "9",
//              "OwnerID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "71559f6b4713193ee97eab26539723b128e6e48a",
//     +            "Hash": "42506a1207e11054a31520d5d1b6a4c662033c02",
//                  "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:9"
//              }
//          }
// u[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3",
//              "LastObjectSize": "394",
//     -        "ModTime": "6",
//     +        "ModTime": "9",
//              "OwnerID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "b2a2a2eb2b6c3e55c052f51f345ed136118c31ab",
//     +                "Hash": "d4b3f299aad861e685d582ad9bbd5aad2f30063f",
//                      "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8"
//                  },
//                  "Index": "0",
// u[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:2](0)=
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "f6a073a0ddcdc3b0bae98ce001a092517e1d9c49",
//     +                "Hash": "cbd1afc8b9193133f4a2611d4dc584a74b694eda",
//                      "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3"
//                  }
//              },
