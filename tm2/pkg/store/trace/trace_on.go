//go:build gastrace

package trace

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const StoreGasEnabled = true

var out *bufio.Writer // nil when writing to stderr (unbuffered)
var outFile *os.File  // always set

func init() {
	path := os.Getenv("GAS_TRACE")
	if path == "" || path == "1" || path == "true" {
		outFile = os.Stderr
		// No bufio for stderr — crash-safe, traces visible immediately.
	} else {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			panic("GAS_TRACE: " + err.Error())
		}
		outFile = f
		out = bufio.NewWriter(f)
	}
}

func Store(op string, gas int64, key []byte, valLen int, info string) {
	keyHex := hex.EncodeToString(key)
	if len(keyHex) > 160 {
		keyHex = keyHex[:160] + "..."
	}
	keyStr := make([]byte, len(key))
	for i, b := range key {
		if b >= 0x20 && b < 0x7f {
			keyStr[i] = b
		} else {
			keyStr[i] = '.'
		}
	}
	if len(keyStr) > 80 {
		keyStr = append(keyStr[:80], '.', '.', '.')
	}
	var w io.Writer = outFile
	if out != nil {
		w = out
	}
	fmt.Fprintf(w,
		"GAS_STORE op=%-14s gas=%-10d vlen=%-6d info=%-16s key_hex=%s key_str=%s\n",
		op, gas, valLen, info, keyHex, keyStr)
}

func TxStart(mode string, gasWanted int64) {
	var w io.Writer = outFile
	if out != nil {
		w = out
	}
	fmt.Fprintf(w, "GAS_TX_START mode=%s gas_wanted=%d\n", mode, gasWanted)
}

func TxEnd(gasUsed int64) {
	var w io.Writer = outFile
	if out != nil {
		w = out
	}
	fmt.Fprintf(w, "GAS_TX_END gas_used=%d\n", gasUsed)
	flush()
}

func TxEndDebug(gasUsed, totalCharge, totalRefund int64) {
	var w io.Writer = outFile
	if out != nil {
		w = out
	}
	fmt.Fprintf(w, "GAS_TX_END gas_used=%d meter_charges=%d meter_refunds=%d meter_net=%d\n",
		gasUsed, totalCharge, totalRefund, totalCharge-totalRefund)
	flush()
}

func flush() {
	if out != nil {
		out.Flush()
	}
}
