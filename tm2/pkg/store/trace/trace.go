//go:build !gastrace

package trace

const StoreGasEnabled = false

func Store(op string, gas int64, key []byte, valLen int, info string) {}
func Flush()                                                           {}
