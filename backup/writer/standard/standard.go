package standard

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

// NewWriter creates a new standard tx data writer
func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer: writer,
	}
}

func (w *Writer) WriteTxData(data *gnoland.TxWithMetadata) error {
	// Marshal the entire tx data into JSON
	jsonData, err := amino.MarshalJSON(data)
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
