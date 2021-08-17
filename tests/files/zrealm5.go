// PKGPATH: gno.land/r/test
package test

import (
	"github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl"
	"github.com/gnolang/gno/_test/timtadh/data-structures/types"
)

var tree *avl.AvlNode

func init() {
	tree, _ = tree.Put(types.String("key0"), "value0")
}

func main() {
	var updated bool
	tree, updated = tree.Put(types.String("key1"), "value1")
	println(updated, tree.Size())
}

// Output:
// false 2

// Realm:
// c[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:5]=(gno.StructValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   _RealmID: (gno.RealmID) RIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30,
//   NewTime: (uint64) 5,
//   _OwnerNewTime: (uint64) 2,
//   _ModTime: (uint64) 0,
//   _RefCount: (int) 1
//  },
//  Fields: ([]gno.TypedValueImage) (len=5 cap=5) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 63CDE69354F70377B65D4C6BDDBD1D23A8AF7217,
//    ValueImage: (gno.PrimitiveValueImage) (len=4 cap=8) {
//     00000000  6b 65 79 31                                       |key1|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 473287F8298DBA7163A897908958F7C0EAE733E2,
//    ValueImage: (gno.PrimitiveValueImage) (len=6 cap=8) {
//     00000000  76 61 6c 75 65 31                                 |value1|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 6DA88C34BA124C41F977DB66A4FC5C1A951708D2,
//    ValueImage: (gno.PrimitiveValueImage) (len=8 cap=8) {
//     00000000  01 00 00 00 00 00 00 00                           |........|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//      ValueImage: (gno.ValueImage) <nil>
//     }
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//      ValueImage: (gno.ValueImage) <nil>
//     }
//    }
//   }
//  }
// }
//
// u[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:2]=(gno.StructValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   _RealmID: (gno.RealmID) RIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30,
//   NewTime: (uint64) 2,
//   _OwnerNewTime: (uint64) 0,
//   _ModTime: (uint64) 3,
//   _RefCount: (int) 1
//  },
//  Fields: ([]gno.TypedValueImage) (len=5 cap=5) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 63CDE69354F70377B65D4C6BDDBD1D23A8AF7217,
//    ValueImage: (gno.PrimitiveValueImage) (len=4 cap=8) {
//     00000000  6b 65 79 30                                       |key0|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 473287F8298DBA7163A897908958F7C0EAE733E2,
//    ValueImage: (gno.PrimitiveValueImage) (len=6 cap=8) {
//     00000000  76 61 6c 75 65 30                                 |value0|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 6DA88C34BA124C41F977DB66A4FC5C1A951708D2,
//    ValueImage: (gno.PrimitiveValueImage) (len=8 cap=8) {
//     00000000  02 00 00 00 00 00 00 00                           |........|
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 0000000000000000000000000000000000000000,
//      ValueImage: (gno.ValueImage) <nil>
//     }
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 4AF0F175D54357F0FEEAE4CF180A42BE848369E8,
//      ValueImage: (gno.RefImage) {
//       RealmID: (gno.RealmID) RID0000000000000000000000000000000000000000,
//       NewTime: (uint64) 5,
//       Hash: (gno.ValueHash) {
//        Hashlet: (gno.Hashlet) (len=20 cap=20) {
//         00000000  c6 41 47 8a 20 67 38 4e  d8 18 2b 73 67 a6 8b ea  |.AG. g8N..+sg...|
//         00000010  65 a7 45 f7                                       |e.E.|
//        }
//       }
//      }
//     }
//    }
//   }
//  }
// }
//
// u[OIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30:0]=(gno.BlockValueImage) {
//  ObjectInfo: (gno.ObjectInfoImage) {
//   _RealmID: (gno.RealmID) RIDA8ADA09DEE16D791FD406D629FE29BB0ED084A30,
//   NewTime: (uint64) 0,
//   _OwnerNewTime: (uint64) 0,
//   _ModTime: (uint64) 4,
//   _RefCount: (int) 0
//  },
//  ParentID: (gno.ObjectID) OIDNONE:0,
//  Values: ([]gno.TypedValueImage) (len=3 cap=3) {
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) 0BA050DA455A6AAD7074EB2148D53ECD5BECC26D,
//    ValueImage: (gno.FuncValueImage) {
//     TypeID: (gno.TypeID) (len=20 cap=20) 0BA050DA455A6AAD7074EB2148D53ECD5BECC26D,
//     IsMethod: (bool) false,
//     Name: (gno.Name) (len=6) "init.0",
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
//     FileName: (gno.Name) (len=16) "files/zrealm5.go",
//     PkgPath: (string) (len=15) "gno.land/r/test"
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
//     FileName: (gno.Name) (len=16) "files/zrealm5.go",
//     PkgPath: (string) (len=15) "gno.land/r/test"
//    }
//   },
//   (gno.TypedValueImage) {
//    TypeID: (gno.TypeID) (len=20 cap=20) E6E0E2CE563ADB23D6A4822DD5FC346A5DE899A0,
//    ValueImage: (gno.PointerValueImage) {
//     TypedValue: (gno.TypedValueImage) {
//      TypeID: (gno.TypeID) (len=20 cap=20) 4AF0F175D54357F0FEEAE4CF180A42BE848369E8,
//      ValueImage: (gno.RefImage) {
//       RealmID: (gno.RealmID) RID0000000000000000000000000000000000000000,
//       NewTime: (uint64) 2,
//       Hash: (gno.ValueHash) {
//        Hashlet: (gno.Hashlet) (len=20 cap=20) {
//         00000000  c6 91 b1 83 9f a8 8b ba  79 a8 5c 64 de c4 9d 59  |........y.\d...Y|
//         00000010  fc 38 b9 6d                                       |.8.m|
//        }
//       }
//      }
//     }
//    }
//   }
//  }
// }
