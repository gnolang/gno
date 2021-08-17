// PKGPATH: gno.land/r/test
package test

var root Node

type Node interface{}
type Key interface{}

type InnerNode struct {
	Key   Key
	Left  Node
	Right Node
}

func main() {
	key := "somekey"
	root = InnerNode{
		Key:   key,
		Left:  nil,
		Right: nil,
	}
}

// Realm:
// c[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:2]=(gno.StructValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   _RealmID: (gno.RealmID) RIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30,
//   NewTime: (uint64) 2,
//   _OwnerNewTime: (uint64) 0,
//   _ModTime: (uint64) 0,
//   _RefCount: (int) 1
//  },
//  Fields: ([]gno.TypedValueImage) (len=3 cap=3) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 473287F8298DBA7163A897908958F7C0EAE733E2,
//    ValueImage: (gno.PrimitiveValueImage) (len=7 cap=8) {
//     00000000  73 6f 6d 65 6b 65 79                              |somekey|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//    ValueImage: (gno.ValueImage) <nil>
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//    ValueImage: (gno.ValueImage) <nil>
//   }
//  }
// }
//
// u[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:0]=(gno.BlockValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   _RealmID: (gno.RealmID) RIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30,
//   NewTime: (uint64) 0,
//   _OwnerNewTime: (uint64) 0,
//   _ModTime: (uint64) 1,
//   _RefCount: (int) 0
//  },
//  ParentID: (gno.ObjectID) OIDNONE:0,
//  Values: ([]gno.TypedValueImage) (len=5 cap=5) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 1AF40977153D0FABAB9803BF33EDEBA8EB420CC5,
//    ValueImage: (gno.TypeValueImage) {
//     TypeID: (gno.TypeID) (len=20 cap=20) 8F3FCA65F6CA73D096C06F68E24FF93EA462D350
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 1AF40977153D0FABAB9803BF33EDEBA8EB420CC5,
//    ValueImage: (gno.TypeValueImage) {
//     TypeID: (gno.TypeID) (len=20 cap=20) B06B716FF82D41A482D5C1CC3711002B74717639
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 1AF40977153D0FABAB9803BF33EDEBA8EB420CC5,
//    ValueImage: (gno.TypeValueImage) {
//     TypeID: (gno.TypeID) (len=20 cap=20) CE75E799ED699FE6A487D6CA237759F5F203BEE0
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 0BA050DA455A6AAD7074EB2148D53ECD5BECC26D,
//    ValueImage: (gno.FuncValueImage) {
//     TypeID: (gno.TypeID) (len=20 cap=20) 0BA050DA455A6AAD7074EB2148D53ECD5BECC26D,
//     IsMethod: (bool) false,
//     Name: (gno.Name) (len=4) "main",
//     ClosureRef: (gno.RefImage) {
//      RealmID: (gno.RealmID) RID0000000000000000000000000000000000000000,
//      NewTime: (uint64) 0,
//      Hash: (gno.ValueHash) {
//       Hashlet: (gno.Hashlet) (len=20 cap=20) {
//        00000000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
//        00000010  00 00 00 00                                       |....|
//       }
//      }
//     },
//     FileName: (gno.Name) (len=16) "files/zrealm1.go",
//     PkgPath: (string) (len=15) "gno.land/r/test"
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) CE75E799ED699FE6A487D6CA237759F5F203BEE0,
//    ValueImage: (gno.StructValueImage) {
//     ObjectInfo: (gno.ObjectInfoImage) {
//      _RealmID: (gno.RealmID) RIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30,
//      NewTime: (uint64) 2,
//      _OwnerNewTime: (uint64) 0,
//      _ModTime: (uint64) 0,
//      _RefCount: (int) 1
//     },
//     Fields: ([]gno.TypedValueImage) (len=3 cap=3) {
//      (gno.TypedValueImage) {
//       TypeID: (gno.TypeID) (len=20 cap=20) 473287F8298DBA7163A897908958F7C0EAE733E2,
//       ValueImage: (gno.PrimitiveValueImage) (len=7 cap=8) {
//        00000000  73 6f 6d 65 6b 65 79                              |somekey|
//       }
//      },
//      (gno.TypedValueImage) {
//       TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//       ValueImage: (gno.ValueImage) <nil>
//      },
//      (gno.TypedValueImage) {
//       TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//       ValueImage: (gno.ValueImage) <nil>
//      }
//     }
//    }
//   }
//  }
// }
//
