package gnomod

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/toml"
)

// maxDecodeErrBytes bounds the decoder's own error text. go-toml quotes what it
// cannot read into the message: a string in the wrong field makes the text as
// long as the body, and an array under a table header several times it. Every
// caller copies the text whole (sdk.clipLog splits it line by line before capping
// it, so its cap does not save the allocation), and it reaches the transaction's
// log. Nothing reads it programmatically; it is a diagnostic, so truncate it at
// the source.
const maxDecodeErrBytes = 1 << 10

// ParseTomlBytes parses the gnomod.toml file from bytes.
func parseTomlBytes(fname string, data []byte) (*File, error) {
	var f File
	if err := toml.Unmarshal(data, &f); err != nil {
		// Not %w: the wrapped text is what is oversized, and no caller
		// unwraps to a go-toml error.
		return nil, fmt.Errorf("error parsing gnomod.toml file at %q: %s", fname, truncateDecodeErr(err.Error()))
	}
	return &f, nil
}

func truncateDecodeErr(s string) string {
	if len(s) <= maxDecodeErrBytes {
		return s
	}
	return s[:maxDecodeErrBytes] + "...<truncated>"
}

// WriteTomlString writes the gnomod.toml file to a string.
func (f *File) WriteString() string {
	var builder strings.Builder
	encoder := toml.NewEncoder(&builder)
	encoder.Order(toml.OrderPreserve)
	// encoder.PromoteAnonymous(true)

	err := encoder.Encode(f)
	if err != nil {
		panic(err)
	}

	return builder.String()
}
