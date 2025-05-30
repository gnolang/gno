package legacy

//nolint:revive // See https://github.com/gnolang/gno/issues/1197
import (
	"fmt"
	"io"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	_ "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

type Writer struct {
	writer io.Writer
}

// NewWriter creates a new legacy tx writer
func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer: writer,
	}
}

func (w *Writer) WriteTxData(data *gnoland.TxWithMetadata) error {
	// Marshal tx individual tx into JSON, instead of the entire tx data
	jsonData, err := amino.MarshalJSON(data.Tx)
	if err != nil {
		return fmt.Errorf("unable to marshal JSON data, %w", err)
	}

	// Write the JSON data as a line to the file
	_, err = w.writer.Write(jsonData)
	if err != nil {
		return fmt.Errorf("unable to write to output, %w", err)
	}

	// Write a newline character to separate JSON objects
	_, err = w.writer.Write([]byte("\n"))
	if err != nil {
		return fmt.Errorf("unable to write newline output, %w", err)
	}

	return nil
}
