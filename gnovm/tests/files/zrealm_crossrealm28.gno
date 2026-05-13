// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type Struct struct {
	A int
}

var s *Struct

func init() {
	s = &Struct{A: 100}
	crossrealm_b.SetObject(cross, func() { println(&s.A) })
}

func main(cur realm) {
	//s := &Struct{A: 100}
	crossrealm_b.SetObject(cross, 123)
	s.A = 123
	println(s)
}

// Output:
// &(struct{(123 int)} gno.land/r/crossrealm.Struct)

// Preprocessed:
// file{ package crossrealm; import crossrealm_b gno.land/r/tests/vm/crossrealm_b; type Struct (const-type gno.land/r/crossrealm.Struct); var s<!~VPBlock(2,1)> *(typeval{gno.land/r/crossrealm.Struct}); func init.2() { s<~VPBlock(3,1)> = &((const-type gno.land/r/crossrealm.Struct){A<VPField(0,0,A)>: (const (100 int))}); (const (ref(gno.land/r/tests/vm/crossrealm_b) package{})).SetObject((const (undefined)), func func(){ (const (println func(...interface {})))(&((const (ref(gno.land/r/crossrealm) package{})).s.A)) }) }; func main(cur (const-type .uverse.realm)) { (const (ref(gno.land/r/tests/vm/crossrealm_b) package{})).SetObject((const (undefined)), (const (123 int))); s<~VPBlock(3,1)>.A = (const (123 int)); (const (println func(...interface {})))(s<~VPBlock(3,1)>) } }

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[06279d1e03ecb38b84822f9caf584bd16913470d:13](-98)=
//     @@ -2,20 +2,15 @@
//          "ObjectInfo": {
//              "ID": "06279d1e03ecb38b84822f9caf584bd16913470d:13",
//              "LastObjectSize": "320",
//     -        "ModTime": "18",
//     +        "ModTime": "19",
//              "OwnerID": "06279d1e03ecb38b84822f9caf584bd16913470d:2",
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
//     -            "Hash": "8b50fbe1097c0c1a442a7b9da3b4162702c78318",
//     -            "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:19"
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
//     -                "Hash": "49c47df4f68f600f4dafa3596d79e4305ddc80e8",
//     +                "Hash": "cf1b0ac375728467108dd7d0e43198287f192a9d",
//                      "ObjectID": "06279d1e03ecb38b84822f9caf584bd16913470d:13"
//                  }
//              },
// d[06279d1e03ecb38b84822f9caf584bd16913470d:19](-298)
// finalizerealm["gno.land/r/crossrealm"]
// u[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8](5)=
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
//              "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8",
//              "LastObjectSize": "214",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:7",
//              "RefCount": "1"
//          }
// u[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:7](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:7",
//              "LastObjectSize": "344",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "d610ed62030736e386292c49e46dd345731c5729",
//     +            "Hash": "f347385e9ab6884b40a8a409a033ec025bfb90b3",
//                  "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8"
//              }
//          }
// u[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3",
//              "LastObjectSize": "394",
//     -        "ModTime": "6",
//     +        "ModTime": "8",
//              "OwnerID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "3a106f3891007bca32bd8e1da44e4e473d6b77e6",
//     +                "Hash": "2db709b8cda0d2dbddcccebdabea37979d0c76cc",
//                      "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:7"
//                  },
//                  "Index": "0",
// u[0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:2](0)=
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "1d35b03a1d535122642af7b7e366e15e1b0ef43a",
//     +                "Hash": "f507af70ccbde358b3ff294278f913b0ea6b6f50",
//                      "ObjectID": "0c8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3"
//                  }
//              },
