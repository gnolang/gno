//go:build !gastrace

package trace

const StoreGasEnabled = false

func Store(op string, gas int64, key []byte, valLen int, info string) {}
func Flush()                                                           {}
func TxStart(mode string, gasWanted int64)                             {}
func TxEnd(gasUsed int64)                                              {}
