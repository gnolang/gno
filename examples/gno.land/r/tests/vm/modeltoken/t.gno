package modeltoken

import "gno.land/p/demo/tests/modelledger"

var l = modelledger.New()

func Transfer(cur realm, to address, n int64) { l.Move(cur.Previous().Address(), to, n) }
func Ledger() *modelledger.Ledger             { return l }
